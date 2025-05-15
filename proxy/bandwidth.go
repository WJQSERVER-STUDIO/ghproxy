package proxy

import (
	"errors"
	"ghproxy/config"

	"github.com/WJQSERVER-STUDIO/go-utils/limitreader"
	"golang.org/x/time/rate"
)

var (
	bandwidthLimit rate.Limit
	bandwidthBurst rate.Limit
)

func UnDefiendRateStringErrHandle(err error) error {
	if errors.Is(err, &limitreader.UnDefiendRateStringErr{}) {
		logWarning("UnDefiendRateStringErr: %s", err)
		return nil
	}
	return err
}

func SetGlobalRateLimit(cfg *config.Config) error {
	if cfg.RateLimit.BandwidthLimit.Enabled {
		var err error
		var totalLimit rate.Limit
		var totalBurst rate.Limit
		totalLimit, err = limitreader.ParseRate(cfg.RateLimit.BandwidthLimit.TotalLimit)
		if UnDefiendRateStringErrHandle(err) != nil {
			logError("Failed to parse total bandwidth limit: %v", err)
			return err
		}
		totalBurst, err = limitreader.ParseRate(cfg.RateLimit.BandwidthLimit.TotalBurst)
		if UnDefiendRateStringErrHandle(err) != nil {
			logError("Failed to parse total bandwidth burst: %v", err)
			return err
		}
		limitreader.SetGlobalRateLimit(totalLimit, int(totalBurst))
		err = SetBandwidthLimit(cfg)
		if UnDefiendRateStringErrHandle(err) != nil {
			logError("Failed to set bandwidth limit: %v", err)
			return err
		}
	} else {
		limitreader.SetGlobalRateLimit(rate.Inf, 0)
	}
	return nil
}

func SetBandwidthLimit(cfg *config.Config) error {
	var err error
	bandwidthLimit, err = limitreader.ParseRate(cfg.RateLimit.BandwidthLimit.SingleLimit)
	if UnDefiendRateStringErrHandle(err) != nil {
		logError("Failed to parse bandwidth limit: %v", err)
		return err
	}
	bandwidthBurst, err = limitreader.ParseRate(cfg.RateLimit.BandwidthLimit.SingleBurst)
	if UnDefiendRateStringErrHandle(err) != nil {
		logError("Failed to parse bandwidth burst: %v", err)
		return err
	}
	return nil
}
