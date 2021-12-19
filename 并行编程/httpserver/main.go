package main

import (
	"context"
	"errgrup/server"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func main()  {
	add := ":8080"

	sch := make(chan *http.Server)

	g,ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		s,err := server.StartServer(add)
		if err != nil {
			return err
		}
		sch <- s
		return nil
	})

	g.Wait()

	select {

	case server := <-sch:
		server.Shutdown(ctx)
	case  <-ctx.Done():

     }

}
