package server

import (
	"sort"

	"github.com/runzhammer/gamedemo/pkg/protocol"
)

func sortSessions(sessions []protocol.SessionSummary, playerRating int) {
	sort.SliceStable(sessions, func(i, j int) bool {
		left := sessions[i].AverageRating - playerRating
		if left < 0 {
			left = -left
		}
		right := sessions[j].AverageRating - playerRating
		if right < 0 {
			right = -right
		}
		if left != right {
			return left < right
		}
		return sessions[i].CreatedAt.Before(sessions[j].CreatedAt)
	})
}
