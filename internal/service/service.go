package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"marketflow/internal/domain"
	"marketflow/internal/ports"
)

type Service struct {
	priceProcessor *PriceProcessor
	resultCh       chan domain.Price
	ports          []string
	wg             sync.WaitGroup
}

type PriceProcessor struct {
	priceRepo         ports.PriceRepository
	aggrPriceDataRepo ports.AggrPriceDataRepository
}

func GetService(
	priceRepo ports.PriceRepository,
	aggrPriceDataRepo ports.AggrPriceDataRepository,
	ports []string,
) *Service {
	return &Service{
		priceProcessor: &PriceProcessor{
			priceRepo:         priceRepo,
			aggrPriceDataRepo: aggrPriceDataRepo,
		},
		resultCh: make(chan domain.Price),
		ports:    ports,
	}
}

func (s *Service) Start(ctx context.Context) {
	for _, port := range s.ports {
		s.startFanOut(ctx, port)
	}

	s.wg.Add(1)
	go s.fanIn(ctx)
}

func (s *Service) startFanOut(ctx context.Context, port string) {
	for i := 0; i < 5; i++ {
		s.wg.Add(1)
		go worker(i, ctx, &s.wg, port, s.resultCh)
	}
}

func (s *Service) fanIn(ctx context.Context) {
	defer s.wg.Done()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var batchPrices []domain.Price

	for {
		select {
		case <-ctx.Done():
			log.Println("fanIn stopped")
			return
		case price := <-s.resultCh:
			batchPrices = append(batchPrices, price)
		case <-ticker.C:
			if len(batchPrices) > 0 {
				s.processBatch(batchPrices)
				batchPrices = batchPrices[:0]
			} else {
				log.Println("no data for this interval")
			}
		}
	}
}

func (s *Service) processBatch(prices []domain.Price) {
	type stats struct {
		sum   float64
		count int
		min   float64
		max   float64
	}

	data := make(map[string]*stats)

	for _, p := range prices {
		key := p.Symbol + "|" + p.Exchange

		if _, ok := data[key]; !ok {
			data[key] = &stats{
				min: p.Price,
				max: p.Price,
			}
		}

		st := data[key]
		st.sum += p.Price
		st.count++

		if p.Price < st.min {
			st.min = p.Price
		}
		if p.Price > st.max {
			st.max = p.Price
		}
	}

	now := time.Now()

	for key, st := range data {
		parts := strings.Split(key, "|")

		aggr := domain.AggrPriceData{
			PairName:     parts[0],
			Exchange:     parts[1],
			Timestamp:    now,
			AveragePrice: st.sum / float64(st.count),
			MinPrice:     st.min,
			MaxPrice:     st.max,
		}

		if err := s.priceProcessor.aggrPriceDataRepo.AddNewAggregatedPriceData(aggr); err != nil {
			log.Println("failed to save aggregated data:", err)
		}
	}
}

func (s *Service) Stop() {
	// Подождём пока все воркеры завершат работу
	close(s.resultCh)
	s.wg.Wait()
	log.Println("Service stopped")
}

func worker(id int, ctx context.Context, wg *sync.WaitGroup, port string, resultCh chan<- domain.Price) {
	defer wg.Done()

RECONNECT:
	conn, err := net.Dial("tcp", ":"+port)
	if err != nil {
		log.Println("connection failed:", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second): // пробуем снова через 5 секунд
			goto RECONNECT
		}
	}

	log.Printf("Worker %d connected to port %s\n", id, port)
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
			return
		default:
		}

		var price domain.Price
		if err := json.Unmarshal(scanner.Bytes(), &price); err != nil {
			log.Println("invalid json:", err)
			continue
		}
		price.Exchange = port
		resultCh <- price
	}

	_ = conn.Close()
	log.Printf("Worker %d disconnected from port %s\n", id, port)
}

func (s *Service) GetHighestPriceBySymbol(symbol string, timeFilter string) (domain.Price, error) {
	AllAggrPriceData, err := s.priceProcessor.aggrPriceDataRepo.GelAllAggregatedPriceDataBySymbol(symbol)
	if err != nil {
		return domain.Price{}, err
	}
	if timeFilter != "" {
		AllAggrPriceData, err = FilterByDuration(AllAggrPriceData, timeFilter)
		if err != nil {
			return domain.Price{}, err
		}
	}
	data := FindMaxByField(AllAggrPriceData, "MaxPrice")
	var price domain.Price
	price.Symbol = data.PairName
	price.Price = data.MaxPrice
	price.Timestamp = data.Timestamp.Unix()
	return price, nil
}

func (s *Service) GetHighestPriceBySymbolAndExchange(symbol, exchange, timeFilter string) (domain.Price, error) {
	AllAggrPriceData, err := s.priceProcessor.aggrPriceDataRepo.GelAllAggregatedPriceDataBySymbolAndPort(symbol, exchange)
	if err != nil {
		return domain.Price{}, err
	}
	if timeFilter != "" {
		AllAggrPriceData, err = FilterByDuration(AllAggrPriceData, timeFilter)
		if err != nil {
			return domain.Price{}, err
		}
	}
	data := FindMaxByField(AllAggrPriceData, "MaxPrice")
	var price domain.Price
	price.Symbol = data.PairName
	price.Exchange = data.Exchange
	price.Price = data.MaxPrice
	price.Timestamp = data.Timestamp.Unix()
	return price, nil
}

// FilterByDurationUnix фильтрует записи по "длительности" в строке,
// но использует timestamp напрямую
func FilterByDuration(items []domain.AggrPriceData, durationStr string) ([]domain.AggrPriceData, error) {
	// Конвертируем durationStr в секунды через ConvertToTime
	// (предположим, ConvertToTime возвращает time.Duration)
	d := ConvertToTime(durationStr)

	// Текущий момент в Unix time (секунды)
	nowTs := time.Now().Unix()
	cutoffTs := nowTs - int64(d.Seconds())

	filtered := make([]domain.AggrPriceData, 0)

	for _, item := range items {
		itemTs := item.Timestamp.Unix()
		if itemTs >= cutoffTs {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

func ConvertToTime(interval string) time.Duration {
	n, _ := strconv.Atoi(interval[:len(interval)-1])
	tm := interval[len(interval)-1:]
	switch tm {
	case "s":
		return time.Duration(n) * time.Second
	case "m":
		return time.Duration(n) * time.Minute
	default:
		return time.Duration(n) * time.Hour
	}
}

func FindMaxByField(items []domain.AggrPriceData, field string) domain.AggrPriceData {
	maxIdx := -1
	var maxVal float64

	for i, item := range items {
		v := reflect.ValueOf(item)
		f := v.FieldByName(field)
		if !f.IsValid() {
			panic(fmt.Sprintf("Field '%s' not found in AggrPriceData", field))
		}
		if f.Kind() != reflect.Float64 {
			panic(fmt.Sprintf("Field '%s' is not float64", field))
		}

		val := f.Float()
		if maxIdx == -1 || val > maxVal {
			maxIdx = i
			maxVal = val
		}
	}

	return items[maxIdx]
}
