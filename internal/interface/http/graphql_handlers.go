package http

import (
	"errors"

	"github.com/google/uuid"
	"github.com/morimasaki-web/ponsu/internal/interface/graphqlctx"
)

func viewerFromSession(sess sessionData) (graphqlctx.Viewer, error) {
	if sess.UserID == "" || sess.OrgID == "" {
		return graphqlctx.Viewer{}, errors.New("missing user/org in session")
	}

	userID, err := uuid.Parse(sess.UserID)
	if err != nil {
		return graphqlctx.Viewer{}, errors.New("invalid user_id")
	}
	orgID, err := uuid.Parse(sess.OrgID)
	if err != nil {
		return graphqlctx.Viewer{}, errors.New("invalid org_id")
	}

	return graphqlctx.Viewer{
		UserID: userID,
		OrgID:  orgID,
		Role:   sess.Role,
		Name:   sess.Name,
		Email:  sess.Email,
	}, nil
}
