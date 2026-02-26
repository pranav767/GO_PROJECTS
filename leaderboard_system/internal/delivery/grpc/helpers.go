package grpc

import (
	pb "leaderboard_system/api/gen/leaderboard/v1"
	"leaderboard_system/internal/domain"
)

// toProtoEntries converts domain leaderboard entries to protobuf entries.
func toProtoEntries(entries []domain.LeaderboardEntry) []*pb.LeaderboardEntryProto {
	result := make([]*pb.LeaderboardEntryProto, 0, len(entries))
	for _, e := range entries {
		result = append(result, &pb.LeaderboardEntryProto{
			UserId:   e.UserID,
			Username: e.Username,
			Score:    e.Score,
			Rank:     e.Rank,
		})
	}
	return result
}
