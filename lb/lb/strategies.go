package lb

import (
	"github.com/coldog/proxy/lb/ctx"
	"hash/fnv"
	"math/rand"
)

type Strategy func(h *Handler, c *ctx.Context) *Target

var strategies = map[string]Strategy{
	"wrr":     WRRStrategy,
	"wrrh":    WRRByHealthStrategy,
	"rr":      RRStrategy,
	"rand":    RandStrategy,
	"ip_hash": IPHashStrategy,
}

func Use(name string, dispatcher Strategy) {
	strategies[name] = dispatcher
}

func WRRByHealthStrategy(h *Handler, c *ctx.Context) *Target {
	if len(h.Targets) == 0 {
		return nil
	}

	for _, t := range h.Targets {
		if t.errors == 0 {
			t.Weight = 100
			break
		}

		t.Weight = int((float64(t.errors) / float64(t.requests)) * 100)
	}

	return WRRStrategy(h, c)
}

func WRRStrategy(h *Handler, c *ctx.Context) *Target {
	if len(h.Targets) == 0 {
		return nil
	}

	max, gcd := nums(h)

	for {
		h.index = (h.index + 1) % len(h.Targets)

		if h.index == 0 {
			h.currentWeight = h.currentWeight - gcd
			if h.currentWeight <= 0 {
				h.currentWeight = max
				if h.currentWeight == 0 {
					return nil
				}
			}
		}

		t := h.Targets[h.index]
		if t.Weight >= h.currentWeight {
			return t
		}
	}

	return nil
}

func RandStrategy(h *Handler, c *ctx.Context) *Target {
	pick := rand.Intn(len(h.Targets) - 1)
	h.index = pick
	return h.Targets[pick]
}

func RRStrategy(h *Handler, c *ctx.Context) *Target {
	this := h.index + 1
	if this > len(h.Targets)-1 {
		this = 0
	}

	h.index = this
	return h.Targets[this]
}

func IPHashStrategy(handler *Handler, c *ctx.Context) *Target {
	ip := c.ClientIp()

	if ip == "" || ip == "unknown" {
		return handler.Targets[rand.Intn(len(handler.Targets)-1)]
	}

	h := fnv.New64a()
	h.Write([]byte(ip))
	key := h.Sum64()

	var b, j int64
	for j < int64(len(handler.Targets)-1) {
		b = j
		key = key*2862933555777941757 + 1
		j = int64(float64(b+1) * (float64(int64(1)<<31) / float64((key>>33)+1)))
	}

	handler.index = int(b)
	return handler.Targets[int(b)]
}

func nums(h *Handler) (max, gcd int) {
	for _, t := range h.Targets {
		if t.Weight > max {
			max = t.Weight
		}
	}

	SELECTION:
	for i := 1; i < max; i++ {
		for _, t := range h.Targets {
			if t.Weight % i != 0 {
				continue SELECTION
			}
		}

		if i > gcd {
			gcd = i
		}
	}

	return
}
