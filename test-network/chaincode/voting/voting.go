package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// smartcontract for voting
type SmartContract struct {
	contractapi.Contract
}

// vote structure
type Vote struct {
	Candidate string `json:"candidate"`
	VoterID   string `json:"voterID"`
}

// Initledger initializes the chaincode
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// CastVote allows a voter to cast their vote for a candidate
func (s *SmartContract) CastVote(ctx contractapi.TransactionContextInterface, voterID string, candidate string) error {
	//input validation
	voterID = strings.TrimSpace(voterID)
	if len(voterID) == 0 {

		return fmt.Errorf("voter Id cannot be empty")
	}
	candidate = strings.TrimSpace(candidate)
	if len(candidate) == 0 {

		return fmt.Errorf("candidate name cannot be empty")
	}

	exists, err := s.VoteExists(ctx, voterID)
	if err != nil {
		return fmt.Errorf("failed to check if voter has already voted : %v", err)
	}
	if exists {
		return fmt.Errorf("voter %s has already cast a vote", voterID)
	}

	vote := Vote{
		Candidate: candidate,
		VoterID:   voterID,
	}

	//store the vote
	voteJSON, err := json.Marshal(vote)

	if err != nil {
		return fmt.Errorf("failed to marshal vote data : %v", err)
	}

	err = ctx.GetStub().PutState(voterID, voteJSON)

	if err != nil {
		return fmt.Errorf("failed to store vote%v", err)
	}
	return nil
}

// VoteExists checks if a voter has already cast a vote
func (s *SmartContract) VoteExists(ctx contractapi.TransactionContextInterface, voterID string) (bool, error) {
	voteJSON, err := ctx.GetStub().GetState(voterID)

	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return voteJSON != nil, nil
}

// GetVoteCount returns the total votes per candidate
func (s *SmartContract) GetVoteCount(ctx contractapi.TransactionContextInterface) (map[string]int, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, fmt.Errorf("failed to read get state range: %v", err)
	}
	defer resultsIterator.Close()

	voteCounts := make(map[string]int) // Initialize vote counts for each candidate

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, fmt.Errorf("failed to iterate through results: %v", err)
		}

		var vote Vote

		err = json.Unmarshal(queryResponse.Value, &vote)

		if err != nil {
			//log the error but continue counting the other votes
			log.Printf("Warning: failed to unmarshal vote data: %v", err)
			continue
		}

		voteCounts[vote.Candidate]++
	}
	return voteCounts, nil
}

// GetVote retrieves a specific vote by voterID
func (s *SmartContract) GetVote(ctx contractapi.TransactionContextInterface, voterID string) (*Vote, error) {
	voteJSON, err := ctx.GetStub().GetState(voterID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if voteJSON == nil {
		return nil, fmt.Errorf("the vote for voter %s does not exist", voterID)
	}
	var vote Vote
	err = json.Unmarshal(voteJSON, &vote)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal vote data: %v", err)
	}
	return &vote, nil
}
func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Fatalf("Error creating chaincode: %v", err)
	}
	if err := chaincode.Start(); err != nil {
		log.Fatalf("Error starting chaincode: %v", err)
	}

}
