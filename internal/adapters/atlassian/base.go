package atlassian

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	adagile "github.com/ctreminiom/go-atlassian/jira/agile"
	adajira "github.com/ctreminiom/go-atlassian/jira/v3"
	jiramodels "github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/go-faster/errors"

	appconfig "github.com/kriuchkov/jiraforge/internal/config"
	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

type Gateway struct {
	jiraClient  *adajira.Client
	agileClient *adagile.Client
}

const (
	defaultHTTPTimeout           = 30 * time.Second
	defaultDialTimeout           = 10 * time.Second
	defaultKeepAlive             = 30 * time.Second
	defaultTLSHandshakeTimeout   = 10 * time.Second
	defaultResponseHeaderTimeout = 30 * time.Second
	defaultExpectContinueTimeout = 1 * time.Second
	defaultIdleConnTimeout       = 90 * time.Second
	defaultMaxIdleConns          = 100
	defaultMaxIdleConnsPerHost   = 10
)

func NewGateway(cfg appconfig.AtlassianConfig) (ports.JiraGateway, error) {
	if err := cfg.Validate(); err != nil {
		return nil, errors.Wrap(err, "validate Atlassian config")
	}

	httpClient, err := newHTTPClient(cfg.ProxyURL)
	if err != nil {
		return nil, errors.Wrap(err, "create HTTP client")
	}

	jiraClient, err := adajira.New(httpClient, cfg.Host)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.NewGateway", "create jira client", err)
	}
	jiraClient.Auth.SetBasicAuth(cfg.Email, cfg.Token)

	agileClient, err := adagile.New(httpClient, cfg.Host)
	if err != nil {
		return nil, coreerrors.Wrap(coreerrors.CodeInternal, "adapters.atlassian.NewGateway", "create agile client", err)
	}
	agileClient.Auth.SetBasicAuth(cfg.Email, cfg.Token)

	return &Gateway{jiraClient: jiraClient, agileClient: agileClient}, nil
}

func newHTTPClient(proxyURL string) (*http.Client, error) {
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultDialTimeout,
			KeepAlive: defaultKeepAlive,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          defaultMaxIdleConns,
		MaxIdleConnsPerHost:   defaultMaxIdleConnsPerHost,
		IdleConnTimeout:       defaultIdleConnTimeout,
		TLSHandshakeTimeout:   defaultTLSHandshakeTimeout,
		ResponseHeaderTimeout: defaultResponseHeaderTimeout,
		ExpectContinueTimeout: defaultExpectContinueTimeout,
	}
	if strings.TrimSpace(proxyURL) != "" {
		proxy, err := url.Parse(proxyURL)
		if err != nil {
			return nil, coreerrors.Wrap(coreerrors.CodeInvalidArgument, "adapters.atlassian.newHTTPClient", "parse proxy URL", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	return &http.Client{Transport: transport, Timeout: defaultHTTPTimeout}, nil
}

func (g *Gateway) resolveBoardIDs(ctx context.Context, boardID, projectKey string) ([]int, error) {
	if strings.TrimSpace(boardID) != "" {
		value, err := strconv.Atoi(boardID)
		if err != nil {
			return nil, coreerrors.Wrap(coreerrors.CodeInvalidArgument, "adapters.atlassian.resolveBoardIDs", "parse board_id", err)
		}
		return []int{value}, nil
	}

	boards, response, err := g.agileClient.Board.Gets(ctx, &jiramodels.GetBoardsOptions{ProjectKeyOrID: projectKey}, 0, 50)
	if err != nil {
		return nil, wrapAgileError("adapters.atlassian.resolveBoardIDs", "list boards", response, err)
	}
	if len(boards.Values) == 0 {
		return nil, coreerrors.New(coreerrors.CodeNotFound, "adapters.atlassian.resolveBoardIDs", fmt.Sprintf("no boards found for project %s", projectKey))
	}

	result := make([]int, 0, len(boards.Values))
	for _, board := range boards.Values {
		result = append(result, board.ID)
	}
	return result, nil
}

func wrapJiraError(op, action string, response *jiramodels.ResponseScheme, err error) error {
	if response == nil {
		return coreerrors.Wrap(coreerrors.CodeInternal, op, action, err)
	}
	message := action
	if response.Endpoint != "" {
		message = fmt.Sprintf("%s (endpoint: %s)", action, response.Endpoint)
	}
	if strings.TrimSpace(response.Bytes.String()) != "" {
		message = fmt.Sprintf("%s: %s", message, response.Bytes.String())
	}
	return coreerrors.Wrap(mapStatusCode(response.Code), op, message, err)
}

func wrapAgileError(op, action string, response *jiramodels.ResponseScheme, err error) error {
	return wrapJiraError(op, action, response, err)
}

func mapStatusCode(code int) coreerrors.Code {
	switch code {
	case http.StatusBadRequest:
		return coreerrors.CodeInvalidArgument
	case http.StatusUnauthorized:
		return coreerrors.CodeUnauthenticated
	case http.StatusForbidden:
		return coreerrors.CodePermissionDenied
	case http.StatusNotFound:
		return coreerrors.CodeNotFound
	case http.StatusConflict:
		return coreerrors.CodeConflict
	case http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return coreerrors.CodeUnavailable
	default:
		return coreerrors.CodeInternal
	}
}

func sprintScope(boardID, projectKey string) string {
	if strings.TrimSpace(boardID) != "" {
		return fmt.Sprintf("board:%s", boardID)
	}
	return fmt.Sprintf("project:%s", projectKey)
}

func formatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339)
}

func formatHistoryDate(value string) string {
	parsed, err := time.Parse("2006-01-02T15:04:05.999-0700", value)
	if err != nil {
		return value
	}
	return parsed.Format("2006-01-02 15:04:05")
}

func defaultEmpty(value string) string {
	if strings.TrimSpace(value) == "" {
		return "(empty)"
	}
	return value
}

func parseTimeSpent(value string) (int, error) {
	trimmed := strings.ReplaceAll(strings.TrimSpace(value), " ", "")
	if seconds, err := strconv.Atoi(trimmed); err == nil {
		return seconds, nil
	}
	duration, err := time.ParseDuration(trimmed)
	if err != nil {
		return 0, errors.Wrap(err, "parse time spent duration")
	}
	return int(duration.Seconds()), nil
}
