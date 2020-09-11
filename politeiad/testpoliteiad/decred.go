// Copyright (c) 2017-2019 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package testpoliteiad

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	hdfchain "github.com/hdfchain/politeia/decredplugin"
	v1 "github.com/hdfchain/politeia/politeiad/api/v1"
)

const (
	bestBlock uint32 = 1000
)

func (p *TestPoliteiad) authorizeVote(payload string) (string, error) {
	av, err := hdfchain.DecodeAuthorizeVote([]byte(payload))
	if err != nil {
		return "", err
	}

	// Sign authorize vote
	s := p.identity.SignMessage([]byte(av.Signature))
	av.Receipt = hex.EncodeToString(s[:])
	av.Timestamp = time.Now().Unix()
	av.Version = hdfchain.VersionAuthorizeVote

	p.Lock()
	defer p.Unlock()

	// Store authorize vote
	_, ok := p.authorizeVotes[av.Token]
	if !ok {
		p.authorizeVotes[av.Token] = make(map[string]hdfchain.AuthorizeVote)
	}

	r, err := p.record(av.Token)
	if err != nil {
		return "", err
	}

	p.authorizeVotes[av.Token][r.Version] = *av

	// Prepare reply
	avrb, err := hdfchain.EncodeAuthorizeVoteReply(
		hdfchain.AuthorizeVoteReply{
			Action:        av.Action,
			RecordVersion: r.Version,
			Receipt:       av.Receipt,
			Timestamp:     av.Timestamp,
		})
	if err != nil {
		return "", err
	}

	return string(avrb), nil
}

func (p *TestPoliteiad) startVote(payload string) (string, error) {
	sv, err := hdfchain.DecodeStartVoteV2([]byte(payload))
	if err != nil {
		return "", err
	}

	p.Lock()
	defer p.Unlock()

	// Store start vote
	sv.Version = hdfchain.VersionStartVote
	p.startVotes[sv.Vote.Token] = *sv

	// Prepare reply
	endHeight := bestBlock + sv.Vote.Duration
	svr := hdfchain.StartVoteReply{
		Version:          hdfchain.VersionStartVoteReply,
		StartBlockHeight: strconv.FormatUint(uint64(bestBlock), 10),
		EndHeight:        strconv.FormatUint(uint64(endHeight), 10),
		EligibleTickets:  []string{},
	}
	svrb, err := hdfchain.EncodeStartVoteReply(svr)
	if err != nil {
		return "", err
	}

	// Store reply
	p.startVoteReplies[sv.Vote.Token] = svr

	return string(svrb), nil
}

func (p *TestPoliteiad) startVoteRunoff(payload string) (string, error) {
	svr, err := hdfchain.DecodeStartVoteRunoff([]byte(payload))
	if err != nil {
		return "", err
	}

	p.Lock()
	defer p.Unlock()

	// Store authorize votes
	avReply := make(map[string]hdfchain.AuthorizeVoteReply)
	for _, av := range svr.AuthorizeVotes {
		r, err := p.record(av.Token)
		if err != nil {
			return "", err
		}
		// Fill client data
		s := p.identity.SignMessage([]byte(av.Signature))
		av.Version = hdfchain.VersionAuthorizeVote
		av.Receipt = hex.EncodeToString(s[:])
		av.Timestamp = time.Now().Unix()
		av.Version = hdfchain.VersionAuthorizeVote

		// Store
		_, ok := p.authorizeVotes[av.Token]
		if !ok {
			p.authorizeVotes[av.Token] = make(map[string]hdfchain.AuthorizeVote)
		}
		p.authorizeVotes[av.Token][r.Version] = av

		// Prepare response
		avr := hdfchain.AuthorizeVoteReply{
			Action:        av.Action,
			RecordVersion: r.Version,
			Receipt:       av.Receipt,
			Timestamp:     av.Timestamp,
		}
		avReply[av.Token] = avr
	}

	// Store start votes
	svReply := hdfchain.StartVoteReply{}
	for _, sv := range svr.StartVotes {
		sv.Version = hdfchain.VersionStartVote
		p.startVotes[sv.Vote.Token] = sv
		// Prepare response
		endHeight := bestBlock + sv.Vote.Duration
		svReply.Version = hdfchain.VersionStartVoteReply
		svReply.StartBlockHeight = strconv.FormatUint(uint64(bestBlock), 10)
		svReply.EndHeight = strconv.FormatUint(uint64(endHeight), 10)
		svReply.EligibleTickets = []string{}
	}

	// Store start vote runoff
	p.startVotesRunoff[svr.Token] = *svr

	response := hdfchain.StartVoteRunoffReply{
		AuthorizeVoteReplies: avReply,
		StartVoteReply:       svReply,
	}

	p.startVotesRunoffReplies[svr.Token] = response

	svrReply, err := hdfchain.EncodeStartVoteRunoffReply(response)
	if err != nil {
		return "", err
	}

	return string(svrReply), nil
}

// decredExec executes the passed in plugin command.
func (p *TestPoliteiad) decredExec(pc v1.PluginCommand) (string, error) {
	switch pc.Command {
	case hdfchain.CmdStartVote:
		return p.startVote(pc.Payload)
	case hdfchain.CmdStartVoteRunoff:
		return p.startVoteRunoff(pc.Payload)
	case hdfchain.CmdAuthorizeVote:
		return p.authorizeVote(pc.Payload)
	case hdfchain.CmdBestBlock:
		return strconv.FormatUint(uint64(bestBlock), 10), nil
	case hdfchain.CmdVoteSummary:
		// This is a cache plugin command. No work needed here.
		return "", nil
	}
	return "", fmt.Errorf("invalid testpoliteiad plugin command")
}
