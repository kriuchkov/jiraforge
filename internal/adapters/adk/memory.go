package adk

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"unicode"

	"google.golang.org/genai"

	"google.golang.org/adk/model"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

const (
	defaultSearchLimit     = 8
	defaultPromptLimit     = 5
	defaultLoadMemoryLimit = 10
	maxHitTextLength       = 800
	maxPromptTextLength    = 400
)

const preloadInstructions = `Relevant snippets from earlier sessions with the same user are included below.
Use them only when they help answer the current request.
Prefer the current conversation if there is any conflict.
<PAST_SESSIONS>
%s
</PAST_SESSIONS>`

type MemoryHit struct {
	SessionID string `json:"session_id"`
	Author    string `json:"author,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Text      string `json:"text"`
}

type SessionMemorySearchRequest struct {
	AppName          string
	UserID           string
	CurrentSessionID string
	Query            string
	Limit            int
}

type SessionMemorySearch struct {
	sessionService session.Service
}

type loadMemoryInput struct {
	Query string `json:"query"`
}

type loadMemoryOutput struct {
	Memories []MemoryHit `json:"memories"`
}

type preloadMemoryTool struct {
	search *SessionMemorySearch
	limit  int
}

type scoredHit struct {
	hit   MemoryHit
	score int
	index int
}

func NewSessionMemorySearch(sessionService session.Service) *SessionMemorySearch {
	return &SessionMemorySearch{sessionService: sessionService}
}

func NewLoadMemoryTool(search *SessionMemorySearch) (tool.Tool, error) {
	if search == nil {
		return nil, fmt.Errorf("session memory search is required")
	}

	return functiontool.New(functiontool.Config{
		Name:        "load_memory",
		Description: "Search relevant snippets from previous sessions for the current user.",
	}, func(ctx tool.Context, input loadMemoryInput) (loadMemoryOutput, error) {
		hits, err := search.Search(ctx, SessionMemorySearchRequest{
			AppName:          ctx.AppName(),
			UserID:           ctx.UserID(),
			CurrentSessionID: ctx.SessionID(),
			Query:            input.Query,
			Limit:            defaultLoadMemoryLimit,
		})
		if err != nil {
			return loadMemoryOutput{}, err
		}
		return loadMemoryOutput{Memories: hits}, nil
	})
}

func NewPreloadMemoryTool(search *SessionMemorySearch) tool.Tool {
	return &preloadMemoryTool{search: search, limit: defaultPromptLimit}
}

func (t *preloadMemoryTool) Name() string {
	return "preload_memory"
}

func (t *preloadMemoryTool) Description() string {
	return "Preload relevant context from previous sessions for the current user."
}

func (t *preloadMemoryTool) IsLongRunning() bool {
	return false
}

func (t *preloadMemoryTool) ProcessRequest(ctx tool.Context, req *model.LLMRequest) error {
	if t.search == nil {
		return nil
	}

	query := firstUserText(ctx.UserContent())
	if strings.TrimSpace(query) == "" {
		return nil
	}

	hits, err := t.search.Search(ctx, SessionMemorySearchRequest{
		AppName:          ctx.AppName(),
		UserID:           ctx.UserID(),
		CurrentSessionID: ctx.SessionID(),
		Query:            query,
		Limit:            t.limit,
	})
	if err != nil {
		return fmt.Errorf("preload memory search: %w", err)
	}
	if len(hits) == 0 {
		return nil
	}

	appendInstructions(req, fmt.Sprintf(preloadInstructions, formatPromptHits(hits)))
	return nil
}

func (s *SessionMemorySearch) Search(ctx context.Context, req SessionMemorySearchRequest) ([]MemoryHit, error) {
	if s == nil || s.sessionService == nil {
		return nil, fmt.Errorf("session memory search is not configured")
	}
	if strings.TrimSpace(req.AppName) == "" || strings.TrimSpace(req.UserID) == "" {
		return nil, nil
	}

	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, nil
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}

	list, err := s.sessionService.List(ctx, &session.ListRequest{AppName: req.AppName, UserID: req.UserID})
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	queryWords := wordSet(query)
	queryLower := strings.ToLower(query)
	var matches []scoredHit

	for _, listedSession := range list.Sessions {
		if listedSession.ID() == req.CurrentSessionID {
			continue
		}

		fullSession, err := s.sessionService.Get(ctx, &session.GetRequest{
			AppName:   req.AppName,
			UserID:    req.UserID,
			SessionID: listedSession.ID(),
		})
		if err != nil {
			return nil, fmt.Errorf("get session %s: %w", listedSession.ID(), err)
		}

		for event := range fullSession.Session.Events().All() {
			text := eventText(event)
			if text == "" {
				continue
			}

			score := scoreText(text, queryLower, queryWords)
			if score == 0 {
				continue
			}

			matches = append(matches, scoredHit{
				hit: MemoryHit{
					SessionID: listedSession.ID(),
					Author:    strings.TrimSpace(event.Author),
					Timestamp: event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
					Text:      truncateText(text, maxHitTextLength),
				},
				score: score,
				index: len(matches),
			})
		}
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		if matches[i].hit.Timestamp != matches[j].hit.Timestamp {
			return matches[i].hit.Timestamp > matches[j].hit.Timestamp
		}
		return matches[i].index < matches[j].index
	})

	if len(matches) > limit {
		matches = matches[:limit]
	}

	hits := make([]MemoryHit, 0, len(matches))
	for _, match := range matches {
		hits = append(hits, match.hit)
	}
	return hits, nil
}

func firstUserText(content *genai.Content) string {
	if content == nil {
		return ""
	}
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		if text := normalizeWhitespace(part.Text); text != "" {
			return text
		}
	}
	return ""
}

func eventText(event *session.Event) string {
	if event == nil || event.LLMResponse.Content == nil {
		return ""
	}
	parts := make([]string, 0, len(event.LLMResponse.Content.Parts))
	for _, part := range event.LLMResponse.Content.Parts {
		if part == nil {
			continue
		}
		if text := normalizeWhitespace(part.Text); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}

func scoreText(text, queryLower string, queryWords map[string]struct{}) int {
	textLower := strings.ToLower(text)
	textWords := wordSet(textLower)
	score := 0
	if queryLower != "" && strings.Contains(textLower, queryLower) {
		score += len(queryWords) + 2
	}
	for word := range queryWords {
		if _, ok := textWords[word]; ok {
			score++
		}
	}
	return score
}

func wordSet(text string) map[string]struct{} {
	set := make(map[string]struct{})
	var token []rune
	flush := func() {
		if len(token) == 0 {
			return
		}
		set[string(token)] = struct{}{}
		token = token[:0]
	}

	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			token = append(token, r)
			continue
		}
		flush()
	}
	flush()
	return set
}

func formatPromptHits(hits []MemoryHit) string {
	lines := make([]string, 0, len(hits)*3)
	for _, hit := range hits {
		if hit.Timestamp != "" {
			lines = append(lines, fmt.Sprintf("Time: %s", hit.Timestamp))
		}
		lines = append(lines, fmt.Sprintf("Session: %s", hit.SessionID))
		if hit.Author != "" {
			lines = append(lines, fmt.Sprintf("%s: %s", hit.Author, truncateText(hit.Text, maxPromptTextLength)))
		} else {
			lines = append(lines, truncateText(hit.Text, maxPromptTextLength))
		}
		lines = append(lines, "")
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func appendInstructions(req *model.LLMRequest, instructions ...string) {
	if len(instructions) == 0 {
		return
	}

	joined := strings.Join(instructions, "\n\n")
	if req.Config == nil {
		req.Config = &genai.GenerateContentConfig{}
	}
	if req.Config.SystemInstruction == nil {
		req.Config.SystemInstruction = genai.NewContentFromText(joined, genai.RoleUser)
		return
	}
	if len(req.Config.SystemInstruction.Parts) > 0 {
		last := req.Config.SystemInstruction.Parts[len(req.Config.SystemInstruction.Parts)-1]
		if last != nil && last.Text != "" {
			last.Text += "\n\n" + joined
			return
		}
	}
	req.Config.SystemInstruction.Parts = append(req.Config.SystemInstruction.Parts, genai.NewPartFromText(joined))
}

func normalizeWhitespace(value string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(value)), " ")
}

func truncateText(value string, max int) string {
	if max <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return strings.TrimSpace(string(runes[:max])) + "..."
}
