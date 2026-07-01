package runtimeapi

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/zhongxinqiao06-spec/DreamWorker/engine/internal/contracts/generated"
)

const EngineVersion = "0.1.0"
const ContractSchemaVersion = "0.1"

type PingResponse = generated.RuntimePingResponse

func NewTraceID() string {
	var suffix [4]byte
	if _, err := rand.Read(suffix[:]); err != nil {
		return "tr_" + time.Now().UTC().Format("20060102T150405000")
	}
	return "tr_" + time.Now().UTC().Format("20060102T150405000") + "_" + hex.EncodeToString(suffix[:])
}

func Ping(traceID string) PingResponse {
	if traceID == "" {
		traceID = NewTraceID()
	}

	return PingResponse{
		SchemaVersion: ContractSchemaVersion,
		OK:            true,
		EngineVersion: EngineVersion,
		TraceID:       traceID,
	}
}
