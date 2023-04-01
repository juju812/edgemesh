package tunnel

import (
	"github.com/libp2p/go-libp2p"
	rcmgr "github.com/libp2p/go-libp2p/p2p/host/resource-manager"

	"github.com/kubeedge/edgemesh/pkg/apis/config/defaults"
	"github.com/kubeedge/edgemesh/pkg/apis/config/v1alpha1"
)

func useLimit(config *v1alpha1.TunnelLimitConfig) rcmgr.LimitConfig {
	scalingLimits := rcmgr.DefaultLimits
	protoLimit := rcmgr.BaseLimit{
		Streams:         config.TunnelBaseStreamIn + config.TunnelBaseStreamOut,
		StreamsInbound:  config.TunnelBaseStreamIn,
		StreamsOutbound: config.TunnelBaseStreamOut,
		FD:              rcmgr.DefaultLimits.ProtocolBaseLimit.FD,
		Memory:          rcmgr.DefaultLimits.ProtocolBaseLimit.Memory,
	}
	scalingLimits.AddProtocolLimit(defaults.ProxyProtocol, protoLimit, rcmgr.DefaultLimits.ProtocolLimitIncrease)
	scalingLimits.ProtocolPeerBaseLimit.Streams = config.TunnelPeerBaseStreamIn + config.TunnelPeerBaseStreamOut
	scalingLimits.ProtocolPeerBaseLimit.StreamsOutbound = config.TunnelPeerBaseStreamOut
	scalingLimits.ProtocolPeerBaseLimit.StreamsInbound = config.TunnelPeerBaseStreamIn

	// Add limits around included libp2p protocols
	libp2p.SetDefaultServiceLimits(&scalingLimits)
	// Turn the scaling limits into a static set of limits using `.AutoScale`. This
	// scales the limits proportional to your system memory.
	limits := scalingLimits.AutoScale()
	return limits
}

func CreateLimitOpt(config *v1alpha1.TunnelLimitConfig) (libp2p.Option, error) {
	var limits rcmgr.LimitConfig
	if config.Enable {
		limits = useLimit(config)
	} else {
		limits = rcmgr.InfiniteLimits
	}

	// The resource manager expects a limiter, se we create one from our limits.
	limiter := rcmgr.NewFixedLimiter(limits)

	// FIXME: add NewStatsTraceReporter to ResourceManager

	// Initialize the resource manager
	rm, err := rcmgr.NewResourceManager(limiter)
	if err != nil {
		return nil, err
	}
	return libp2p.ResourceManager(rm), nil
}
