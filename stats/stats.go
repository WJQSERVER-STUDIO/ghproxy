package stats

import (
	"sync"
	"time"
)

// ProxyStats store one ip's proxy stats
// ProxyStats 存储一个IP的代理统计信息
type ProxyStats struct {
	IP               string    `json:"ip"`
	LastCalled       time.Time `json:"last_called"`
	CallCount        int64     `json:"call_count"`
	TotalTransferred int64     `json:"total_transferred"`
}

var (
	statsMap = &sync.Map{}
)

// Record update a ip's proxy stats
// Record 更新一个IP的代理统计信息
func Record(ip string, transferred int64) {
	s, _ := statsMap.LoadOrStore(ip, &ProxyStats{
		IP: ip,
	})

	ps := s.(*ProxyStats)
	ps.LastCalled = time.Now()
	ps.CallCount++
	ps.TotalTransferred += transferred
	statsMap.Store(ip, ps)
}

// GetStats return all proxy stats
// GetStats 返回所有的代理统计信息
func GetStats() map[string]*ProxyStats {
	data := make(map[string]*ProxyStats)
	statsMap.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(*ProxyStats)
		return true
	})
	return data
}
