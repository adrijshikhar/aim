package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aim-cli/aim/internal/session"
)

var (
	safeIdentifierPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)
	safeFlagPattern       = regexp.MustCompile(`^--?[a-zA-Z0-9_./=-]+$`)
)

type sessionService struct {
	mgr      *session.Manager
	launcher LauncherService
}

// NewSessionService creates a new SessionService wrapping session.Manager and LauncherService.
func NewSessionService(mgr *session.Manager, launcher ...LauncherService) SessionService {
	var l LauncherService
	if len(launcher) > 0 && launcher[0] != nil {
		l = launcher[0]
	} else {
		l = NewLauncherService()
	}
	return &sessionService{
		mgr:      mgr,
		launcher: l,
	}
}

// ListSessions queries sessions from storage, applies filtering, and enriches metadata.
func (s *sessionService) ListSessions(ctx context.Context, filter SessionFilter) ([]SessionDTO, error) {
	if s.mgr == nil {
		return []SessionDTO{}, nil
	}

	mgrSessions, err := s.mgr.ListSessions(ctx, filter.Agent, filter.Profile, false)
	if err != nil {
		return nil, err
	}

	query := strings.ToLower(strings.TrimSpace(filter.Query))
	results := make([]SessionDTO, 0, len(mgrSessions))

	for _, sess := range mgrSessions {
		cwd := sess.Cwd
		if cwd == "" && s.mgr != nil {
			cwd, _ = s.mgr.ResolveCwd(ctx, &sess)
		}

		goal := sess.Goal
		if goal == "" && s.mgr != nil {
			if sum, err := s.mgr.ResolveSummary(ctx, &sess); err == nil {
				goal = sum.Goal
				if goal == "" {
					goal = sum.RecentActivity
				}
			}
		}
		if goal == "" {
			goal = sess.Summary
		}

		if query != "" {
			match := strings.Contains(strings.ToLower(sess.ID), query) ||
				strings.Contains(strings.ToLower(sess.ShortID), query) ||
				strings.Contains(strings.ToLower(sess.Title), query) ||
				strings.Contains(strings.ToLower(goal), query) ||
				strings.Contains(strings.ToLower(cwd), query) ||
				strings.Contains(strings.ToLower(sess.Summary), query) ||
				strings.Contains(strings.ToLower(sess.Agent), query) ||
				strings.Contains(strings.ToLower(sess.Profile), query)
			if !match {
				continue
			}
		}

		dto := SessionDTO{
			ID:        sess.ID,
			Agent:     sess.Agent,
			Profile:   sess.Profile,
			Title:     sess.Title,
			Cwd:       cwd,
			Goal:      goal,
			Turns:     sess.MessageCount,
			UpdatedAt: sess.LastActiveAt,
			IsActive:  sess.Status == session.StatusActive,
			PID:       sess.PID,
		}
		results = append(results, dto)

		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	return results, nil
}

// ResumeSessionInTerminal resolves the target session and profile, validates flags, then launches an interactive terminal window.
func (s *sessionService) ResumeSessionInTerminal(ctx context.Context, req ResumeRequest) error {
	agent := strings.TrimSpace(req.Agent)
	if agent == "" {
		return errors.New("agent is required to resume session")
	}
	if !safeIdentifierPattern.MatchString(agent) {
		return fmt.Errorf("invalid agent %q: contains illegal characters", agent)
	}

	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return errors.New("session ID is required to resume session")
	}
	if !safeIdentifierPattern.MatchString(sessionID) {
		return fmt.Errorf("invalid session ID %q: contains illegal characters", sessionID)
	}
	if s.launcher == nil {
		return errors.New("no launcher service available")
	}

	resolvedID := sessionID
	resolvedProfile := strings.TrimSpace(req.Profile)
	if resolvedProfile != "" && resolvedProfile != "<host>" && !safeIdentifierPattern.MatchString(resolvedProfile) {
		return fmt.Errorf("invalid profile %q: contains illegal characters", resolvedProfile)
	}

	if s.mgr != nil {
		sess, err := s.mgr.ResolveSession(ctx, agent, sessionID)
		if err != nil {
			return err
		}
		if sess != nil {
			resolvedID = sess.ID
			if resolvedProfile == "" && sess.Profile != "" && sess.Profile != "<host>" {
				resolvedProfile = sess.Profile
			}
		}
	}

	if resolvedProfile != "" && resolvedProfile != "<host>" && !safeIdentifierPattern.MatchString(resolvedProfile) {
		return fmt.Errorf("invalid resolved profile %q: contains illegal characters", resolvedProfile)
	}
	if !safeIdentifierPattern.MatchString(resolvedID) {
		return fmt.Errorf("invalid resolved session ID %q: contains illegal characters", resolvedID)
	}

	var cmdStr string
	if resolvedProfile != "" && resolvedProfile != "<host>" {
		cmdStr = fmt.Sprintf("aim resume %s %s %s", agent, resolvedProfile, resolvedID)
	} else {
		cmdStr = fmt.Sprintf("aim resume %s %s", agent, resolvedID)
	}

	var extra []string
	for _, f := range req.Flags {
		f = strings.TrimSpace(f)
		if f != "" {
			if !safeFlagPattern.MatchString(f) {
				return fmt.Errorf("invalid flag %q: contains illegal characters", f)
			}
			extra = append(extra, f)
		}
	}
	if custom := strings.TrimSpace(req.CustomFlags); custom != "" {
		tokens := strings.Fields(custom)
		for _, tok := range tokens {
			if !safeFlagPattern.MatchString(tok) {
				return fmt.Errorf("invalid custom flag token %q: contains illegal characters", tok)
			}
			extra = append(extra, tok)
		}
	}
	if len(extra) > 0 {
		cmdStr += " " + strings.Join(extra, " ")
	}

	// Detach terminal launch context from ephemeral request context
	launchCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.launcher.LaunchTerminal(launchCtx, cmdStr)
}
