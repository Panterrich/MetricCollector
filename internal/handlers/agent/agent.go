package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/jpillora/backoff"
	"github.com/rs/zerolog/log"

	"github.com/Panterrich/MetricCollector/internal/collector"
	"github.com/Panterrich/MetricCollector/pkg/hash"
	"github.com/Panterrich/MetricCollector/pkg/serialization"
)

const MaxAttempts = 3

func ReportAllMetrics(
	ctx context.Context,
	storage collector.Collector,
	client *resty.Client,
	serverAddress string,
	keyHash string) {
	metrics := storage.GetAllMetrics(ctx)
	if len(metrics) == 0 {
		return
	}

	values, err := serialization.ConvertToJSONMetrics(metrics)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	data, err := serialization.MetricsToJSON(values)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}

	hash, err := hash.Message(data, []byte(keyHash))
	if err != nil {
		log.Error().Err(err).Send()
	}

	backoffScheduler := &backoff.Backoff{
		Min:    1 * time.Second,
		Max:    5 * time.Second,
		Factor: 3,
	}

	var resp *resty.Response

	for {
		if ctx.Err() != nil {
			return
		}

		if backoffScheduler.Attempt() == MaxAttempts {
			return
		}

		resp, err = client.R().
			SetBody(data).
			SetPathParams(map[string]string{
				"address":    serverAddress,
				"HashSHA256": string(hash),
			}).Post("http://{address}/updates/")

		if err == nil {
			break
		}

		d := backoffScheduler.Duration()

		log.Info().
			Err(err).
			Dur("time reconnecting", d).
			Send()
		time.Sleep(d)
	}

	fmt.Println(resp, err)
}
