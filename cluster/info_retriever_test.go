package cluster

import (
	"testing"

	"github.com/seanbit/nrpc/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestInfoRetrieverRegion(t *testing.T) {
	t.Parallel()

	c := viper.New()
	c.Set("nrpc.cluster.info.region", "us")
	conf := config.NewViperConfig(c)

	infoRetriever := NewInfoRetriever(*&config.NewConfig(conf).Cluster.Info)

	assert.Equal(t, "us", infoRetriever.Region())
}
