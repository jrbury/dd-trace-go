package elastictraced

import (
	"context"
	"github.com/DataDog/dd-trace-go/tracer"
	"github.com/stretchr/testify/assert"
	elasticv3 "gopkg.in/olivere/elastic.v3"
	elasticv5 "gopkg.in/olivere/elastic.v5"

	"testing"
)

const (
	debug = false
)

func TestClientV5(t *testing.T) {
	assert := assert.New(t)
	testTracer, testTransport := getTestTracer()
	testTracer.DebugLoggingEnabled = debug

	tc := NewTracedHTTPClient("my-es-service", testTracer)
	client, err := elasticv5.NewClient(
		elasticv5.SetURL("http://127.0.0.1:9200"),
		elasticv5.SetHttpClient(tc),
		elasticv5.SetSniff(false),
		elasticv5.SetHealthcheck(false),
	)
	assert.NoError(err)

	_, err = client.Index().
		Index("twitter").Id("1").
		Type("tweet").
		BodyString(`{"user": "test", "message": "hello"}`).
		Do(context.TODO())
	assert.NoError(err)
	checkPUTTrace(assert, testTracer, testTransport)

	_, err = client.Get().Index("twitter").Type("tweet").
		Id("1").Do(context.TODO())
	assert.NoError(err)
	checkGETTrace(assert, testTracer, testTransport)

	_, err = client.Get().Index("not-real-index").
		Id("1").Do(context.TODO())
	assert.Error(err)
	checkErrTrace(assert, testTracer, testTransport)
}

func TestClientV3(t *testing.T) {
	assert := assert.New(t)
	testTracer, testTransport := getTestTracer()
	testTracer.DebugLoggingEnabled = debug

	tc := NewTracedHTTPClient("my-es-service", testTracer)
	client, err := elasticv3.NewClient(
		elasticv3.SetURL("http://127.0.0.1:9201"),
		elasticv3.SetHttpClient(tc),
		elasticv3.SetSniff(false),
		elasticv3.SetHealthcheck(false),
	)
	assert.NoError(err)

	_, err = client.Index().
		Index("twitter").Id("1").
		Type("tweet").
		BodyString(`{"user": "test", "message": "hello"}`).
		DoC(context.TODO())
	assert.NoError(err)
	checkPUTTrace(assert, testTracer, testTransport)

	_, err = client.Get().Index("twitter").Type("tweet").
		Id("1").DoC(context.TODO())
	assert.NoError(err)
	checkGETTrace(assert, testTracer, testTransport)

	_, err = client.Get().Index("not-real-index").
		Id("1").DoC(context.TODO())
	assert.Error(err)
	checkErrTrace(assert, testTracer, testTransport)
}

func checkPUTTrace(assert *assert.Assertions, tracer *tracer.Tracer, transport *tracer.DummyTransport) {
	tracer.FlushTraces()
	traces := transport.Traces()
	assert.Len(traces, 1)
	spans := traces[0]
	assert.Equal("my-es-service", spans[0].Service)
	assert.Equal("PUT /twitter/tweet/?", spans[0].Resource)
	assert.Equal("/twitter/tweet/1", spans[0].GetMeta("elasticsearch.url"))
	assert.Equal("PUT", spans[0].GetMeta("elasticsearch.method"))
}

func checkGETTrace(assert *assert.Assertions, tracer *tracer.Tracer, transport *tracer.DummyTransport) {
	tracer.FlushTraces()
	traces := transport.Traces()
	assert.Len(traces, 1)
	spans := traces[0]
	assert.Equal("my-es-service", spans[0].Service)
	assert.Equal("GET /twitter/tweet/?", spans[0].Resource)
	assert.Equal("/twitter/tweet/1", spans[0].GetMeta("elasticsearch.url"))
	assert.Equal("GET", spans[0].GetMeta("elasticsearch.method"))
}

func checkErrTrace(assert *assert.Assertions, tracer *tracer.Tracer, transport *tracer.DummyTransport) {
	tracer.FlushTraces()
	traces := transport.Traces()
	assert.Len(traces, 1)
	spans := traces[0]
	assert.Equal("my-es-service", spans[0].Service)
	assert.Equal("GET /not-real-index/_all/?", spans[0].Resource)
	assert.Equal("/not-real-index/_all/1", spans[0].GetMeta("elasticsearch.url"))
	assert.NotEmpty(spans[0].GetMeta("error.msg"))
	assert.Equal("*errors.errorString", spans[0].GetMeta("error.type"))
}

// getTestTracer returns a Tracer with a DummyTransport
func getTestTracer() (*tracer.Tracer, *tracer.DummyTransport) {
	transport := &tracer.DummyTransport{}
	tracer := tracer.NewTracerTransport(transport)
	return tracer, transport
}
