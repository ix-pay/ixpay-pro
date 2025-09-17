package snowflake

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/ix-pay/ixpay-pro/internal/config"
)

func SetupSnowflake(cfg *config.Config) (*Snowflake, error) {
	machineID, _ := strconv.ParseInt(cfg.Server.MachineId, 10, 64)
	if machineID < 0 {
		return nil, errors.New("机器码异常")
	}

	sf, err := New(machineID)
	if err != nil {
		return nil, err
	}
	return sf, nil
}

const (
	epoch         int64 = 1609459200000 // 2021-01-01 00:00:00 UTC
	timestampBits uint8 = 41
	machineBits   uint8 = 10
	sequenceBits  uint8 = 12

	maxTimestamp int64 = -1 ^ (-1 << timestampBits)
	maxMachineID int64 = -1 ^ (-1 << machineBits)
	maxSequence  int64 = -1 ^ (-1 << sequenceBits)

	timestampShift = machineBits + sequenceBits
	machineShift   = sequenceBits
)

type Snowflake struct {
	mu        sync.Mutex
	lastStamp int64
	machineID int64
	sequence  int64
}

func New(machineID int64) (*Snowflake, error) {
	if machineID < 0 || machineID > maxMachineID {
		return nil, errors.New("machine ID out of range")
	}
	return &Snowflake{
		lastStamp: 0,
		machineID: machineID,
		sequence:  0,
	}, nil
}

func (s *Snowflake) Generate() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	current := time.Now().UnixMilli()
	timestamp := current - epoch

	if timestamp > maxTimestamp {
		panic("时间戳溢出")
	}

	if current < s.lastStamp {
		panic("时钟倒转了")
	}

	if current == s.lastStamp {
		s.sequence = (s.sequence + 1) & maxSequence
		if s.sequence == 0 {
			for current <= s.lastStamp {
				current = time.Now().UnixMilli()
			}
		}
	} else {
		s.sequence = 0
	}

	s.lastStamp = current
	return (timestamp << timestampShift) |
		(s.machineID << machineShift) |
		s.sequence
}
