package quotes

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

type bitfaker struct {
	mu           sync.RWMutex
	outbox       chan<- TradeEvent
	markets      []Market
	period       time.Duration
	tradeSampler *tradeSampler
}

func newBitfaker(config Config, outbox chan<- TradeEvent) *bitfaker {
	return &bitfaker{
		outbox:       outbox,
		markets:      make([]Market, 0),
		period:       5 * time.Second,
		tradeSampler: newTradeSampler(config.TradeSampler),
	}
}

func (b *bitfaker) Start(markets []Market) error {
	if len(markets) == 0 {
		return errors.New("no markets specified")
	}

	for _, m := range markets {
		if err := b.Subscribe(m); err != nil {
			return err
		}
	}

	go func() {
		for {
			b.mu.RLock()
			for _, v := range b.markets {
				b.createTradeEvent(v)
			}
			b.mu.RUnlock()
			<-time.After(b.period)
		}
	}()
	return nil
}

func (b *bitfaker) Subscribe(market Market) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.markets = append(b.markets, market)
	return nil
}

func (b *bitfaker) Unsubscribe(market Market) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	index := -1
	for i, m := range b.markets {
		if market == m {
			index = i
			break
		}
	}

	if index == -1 {
		return fmt.Errorf("market %s not found", market)
	}

	b.markets = append(b.markets[:index], b.markets[index+1:]...)
	return nil
}

func (b *bitfaker) createTradeEvent(market Market) {
	tr := TradeEvent{
		Market: market.BaseUnit + market.QuoteUnit,
		Price:  decimal.NewFromFloat(2.213),
		Source: DriverBitfaker,
	}

	b.outbox <- tr
}

func (b *bitfaker) Stop() error {
	return nil
}
