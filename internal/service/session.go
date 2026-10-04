package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aim-cli/aim/internal/session"
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
		if query != "" {
			match := strings.Contains(strings.ToLower(sess.ID), query) ||
				strings.Contains(strings.ToLower(sess.ShortID), query) ||
				strings.Contains(strings.ToLower(sess.Title), query) ||
				strings.Contains(strings.ToLower(sess.Goal), query) ||
				strings.Contains(strings.ToLower(sess.Cwd), query) ||
				strings.Contains(strings.ToLower(sess.Summary), query) ||
				strings.Contains(strings.ToLower(sess.Agent), query) ||
				strings.Contains(strings.ToLower(sess.Profile), query)
			if !match {
				continue
			}
		}

		cwd := sess.Cwd
		if cwd == "" {
			cwd, _ = s.mgr.ResolveCwd(ctx, &sess)
		}

		goal := sess.Goal
		if goal == "" {
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
		}
		results = append(results, dto)

		if filter.Limit > 0 && len(results) >= filter.Limit {
			break
		}
	}

	return results, nil
}

// ResumeSessionInTerminal resolves the target session and profile, then launches an interactive terminal window.
func (s *sessionService) ResumeSessionInTerminal(ctx context.Context, req ResumeRequest) error {
	agent := strings.TrimSpace(req.Agent)
	if agent == "" {
		return errors.New("agent is required to resume session")
	}
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		return errors.New("session ID is required to resume session")
	}
	if s.launcher == nil {
		return errors.New("no launcher service available")
	}

	resolvedID := sessionID
	resolvedProfile := strings.TrimSpace(req.Profile)

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

	var cmdStr string
	if resolvedProfile != "" && resolvedProfile != "<host>" {
		cmdStr = fmt.Sprintf("aim resume %s %s %s", agent, resolvedProfile, resolvedID)
	} else {
		cmdStr = fmt.Sprintf("aim resume %s %s", agent, resolvedID)
	}

	return s.launcher.LaunchTerminal(ctx, cmdStr)
}
