package dynamodb

import (
	"fmt"
	"strings"
)

type betsItem struct {
	SK   string `dynamodbav:"SK"`
	Bets []byte `dynamodbav:"b"`
}

func (item betsItem) adminNodeID() (string, error) {
	adminNodeID, _, found := strings.Cut(item.SK, "#")
	if !found || adminNodeID == "" {
		return "", fmt.Errorf("bets item SK %q has no admin node id", item.SK)
	}
	return adminNodeID, nil
}
