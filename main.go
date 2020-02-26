package main

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/pat"
	"github.com/nats-io/nats.go"
	pb "github.com/victorolegovich/news-storage-service/proto"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		println("failed to create a logger zap")
		return
	}

	router := pat.New()
	router.Get("/news/{id:[0-9A-z]+}", getNewsItemH)
	http.Handle("/", router)

	if err = http.ListenAndServe("localhost:8080", nil); err != nil {
		logger.Info("failed to start the server", zap.Error(err))
		return
	}
}

func getNewsItemH(w http.ResponseWriter, r *http.Request) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		println("failed to create a logger zap")

		w.WriteHeader(http.StatusInternalServerError)

		if _, err = w.Write([]byte("There were some internal server errors")); err != nil {
			logger.Error("failed to record the client's response", zap.Error(err))
		}

		return
	}

	broker, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		logger.Error("Nats server connection error", zap.Error(err))

		w.WriteHeader(http.StatusInternalServerError)

		if _, err = w.Write([]byte("There were some internal server errors")); err != nil {
			logger.Error("failed to record the client's response", zap.Error(err))
		}
		return
	}

	id := r.URL.Query().Get(":id")

	msg, err := broker.Request("storage", []byte(id), time.Second)
	if err != nil {
		logger.Error("message sending error news storage service", zap.Error(err))

		w.WriteHeader(http.StatusInternalServerError)

		if _, err = w.Write([]byte("There were some internal server errors")); err != nil {
			logger.Error("failed to record the client's response", zap.Error(err))
		}

		return
	}

	logger.Info("reply to the request was received", zap.String("subject", msg.Subject))

	newsItem := &pb.NewsItem{}

	if err = proto.Unmarshal(msg.Data, newsItem); err != nil {
		logger.Info("The news with this ID could not be found in the repository.", zap.String("id", id))

		w.WriteHeader(http.StatusNotFound)

		if _, err = w.Write(msg.Data); err != nil {
			logger.Error("failed to record the client's response", zap.Error(err))
		}

		return
	}

	marshaller := jsonpb.Marshaler{
		Indent:   "\t",
		OrigName: true,
	}

	if err = marshaller.Marshal(w, newsItem); err != nil {
		logger.Error("protobuf message conversion error to json string")
	}

	w.Header().Set("Content-Type", "application/json;charset=utf-8")

	return
}
