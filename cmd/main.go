package main

import (
	"flag"
	. "github.com/neuegram/gopenmpi/pkg/ompi"
	"log"
)

func main() {
	var hostfile string
	flag.StringVar(&hostfile, "hostfile", "host_file", "newline delimited list of MPI hosts (workers)")
	flag.Parse()

	mpi, err := NewMPI(hostfile)
	if err != nil {
		log.Fatal(err)
	}

	mpi.Initialize()

	msg := []byte("The quick red fox jumps over the lazy dog")
	if err := mpi.Send(&msg, len(msg), 0, 0); err != nil {
		log.Fatal(err)
	}
}

