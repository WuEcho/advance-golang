package server

import (
	"context"
	"net/http"
	"github.com/pkg/errors"
)

func StartServer(address string) (*http.Server,error) {
	s := &http.Server{Addr: address}
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		//do sth
	})
	err := s.ListenAndServe()
	if err != nil {
		return nil,errors.Wrapf(err,"listen and server error :%s address: %s",err.Error(),address)
	}
	return s,nil
}

func Stop(s *http.Server,ctx context.Context) error {
    return s.Shutdown(ctx)
}
