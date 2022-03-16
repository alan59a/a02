package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"
)

type IDX struct {
	data interface{} // It can be either []uint8, []int8, []float32 or []float64 ... gonna change it for 1.18 ... don't worry!
	dims []int
}

func main() {
	a := Read("./example/train-images.idx3-ubyte")
	fmt.Println(a.dims)
}

func Read(addr string) *IDX {
	file, err := os.Open(addr)
	check(err)
	defer file.Close()

	// Checking the file ... the 1st 2 byte are always 0
	var all int = 1
	magic := make([]byte, 4)
	_, err = file.Read(magic)
	check(err)
	if magic[0] != 0 || magic[1] != 0 {
		log.Fatalln("Unknown file compostion ... are you sure this is an IDX file?")
	}
	b4 := make([]byte, 4)

	// The 4th byte is the number of dimensions
	dims := make([]int, int(magic[3]))
	for a := 0; a < int(magic[3]); a++ {
		_, err = file.Read(b4)
		check(err)
		dims[a] = int(binary.BigEndian.Uint32(b4))
		all *= int(binary.BigEndian.Uint32(b4))
	}

	// and it's always nice to have a progress bar
	bar := newBar(int(all), "Loading the file ...")

	// The 3rd byte of an IDX file indicates the data type. Since the file type of the data is unknown ...
	// Remember each data type may need different amounts of bytes
	switch magic[2] {
	case 8:
		// uint8 needs a single byte
		b1 := make([]byte, 1)
		data := make([]uint8, all)

		a := 0
		b := all / 1000
		for {
			_, err = file.Read(b1)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
			// in Golang byte and uint8 are considered the same ... so no conversion is needed in this case
			data[a] = b1[0]
			a++
			b--
			if b == 0 {
				b = all / 1000
				bar.add(b)
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 9:
		// int8 needs a single byte
		b1 := make([]byte, 1)
		data := make([]int8, all)

		a := 0
		for {
			_, err = file.Read(b1)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
			data[a] = int8(b1[0])
			a++
			bar.add(1)
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 11:
		// int16 needs 2 bytes
		b2 := make([]byte, 2)
		data := make([]uint16, all)

		a := 0
		for {
			_, err = file.Read(b2)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
			data[a] = binary.BigEndian.Uint16(b2)
			a++
			bar.add(1)
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 12:
		// int32 needs 4 bytes
		data := make([]uint32, all)

		a := 0
		for {
			_, err = file.Read(b4)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
			data[a] = binary.BigEndian.Uint32(b4)
			a++
			bar.add(1)
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 13:
		// float32 needs 4 bytes
		data := make([]float32, all)

		a := 0
		for {
			_, err = file.Read(b4)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
			data[a] = math.Float32frombits(binary.BigEndian.Uint32(b4))
			a++
			bar.add(1)
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 14:
		// float64 needs 8 bytes
		b8 := make([]byte, 8)
		data := make([]float64, all)

		a := 0
		for {
			_, err = file.Read(b8)
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
			data[a] = math.Float64frombits(binary.BigEndian.Uint64(b8))
			a++
			bar.add(1)
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	default:
		log.Fatalln("Unknown file data type ... are you sure this is an IDX file?")
	}
	return &IDX{}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

// Tried to go for concurrency ... failed and it was very slow ... try later maybe
// You can see the code i tried here ...
/*
package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"math"
	"os"

	"golang.org/x/sync/semaphore"
)

type IDX struct {
	data interface{} // It can be either []uint8, []int8, []float32 or []float64
	dims []int
}

func main() {
	a := Read("./files/train-images.idx3-ubyte")
	fmt.Println(a.dims)
}

func Read(addr string) *IDX {
	file, err := os.Open(addr)
	check(err)
	defer file.Close()

	// Checking the file ... the 1st 2 byte are always 0
	var all int = 1
	magic := make([]byte, 4)
	_, err = file.Read(magic)
	check(err)
	if magic[0] != 0 || magic[1] != 0 {
		log.Fatalln("Unknown file compostion ... are you sure this is an IDX file?")
	}
	b4 := make([]byte, 4)

	// The 4th byte is the number of dimensions
	dims := make([]int, int(magic[3]))
	for a := 0; a < int(magic[3]); a++ {
		_, err = file.Read(b4)
		check(err)
		dims[a] = int(binary.BigEndian.Uint32(b4))
		all *= int(binary.BigEndian.Uint32(b4))
	}
	// Let's make it faster ... concurrently!
	sem := semaphore.NewWeighted(16)

	// The 3rd byte of an IDX file indicates the data type. Since the file type of the data is unknown ...
	// Remember each data type may need different amounts of bytes
	bar := newBar(int(all), "Loading the file ...")
	switch magic[2] {
	case 8:
		// uint8 needs a single byte
		b1 := make([]byte, 1)
		data := make([]uint8, all)

		// and it's always nice to have a progress bar
		a := 0
		b := all / 1000
		for {
			err = sem.Acquire(context.Background(), 1)
			check(err)
			go func() {
				defer func() {
					b--
					if b == 0 {
						b = all / 1000
						bar.add(b)
					}
					sem.Release(1)
				}()
				_, err = file.Read(b1)
				// Yeah ... it's wierd for me either ...
				if err == nil {
					// in Golang byte and uint8 are considered the same ... so no conversion is needed in this case
					data[a] = b1[0]
					a++
				}
			}()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 9:
		// int8 needs a single byte
		b1 := make([]byte, 1)
		data := make([]int8, all)

		a := 0
		for {
			err = sem.Acquire(context.Background(), 1)
			check(err)
			go func() {
				defer func() {
					bar.add(1)
					sem.Release(1)
				}()
				_, err = file.Read(b1)
				if err == nil {
					data[a] = int8(b1[0])
					a++
				}

			}()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 11:
		// int16 needs 2 bytes
		b2 := make([]byte, 2)
		data := make([]uint16, all)

		a := 0
		for {
			err = sem.Acquire(context.Background(), 1)
			check(err)
			go func() {
				defer func() {
					bar.add(1)
					sem.Release(1)
				}()
				_, err = file.Read(b2)
				if err == nil {
					data[a] = binary.BigEndian.Uint16(b2)
					a++
				}
			}()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 12:
		// int32 needs 4 bytes
		data := make([]uint32, all)

		a := 0
		for {
			err = sem.Acquire(context.Background(), 1)
			check(err)
			go func() {
				defer func() {
					bar.add(1)
					sem.Release(1)
				}()
				_, err = file.Read(b4)
				if err == nil {
					data[a] = binary.BigEndian.Uint32(b4)
					a++
				}
			}()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 13:
		// float32 needs 4 bytes
		data := make([]float32, all)

		a := 0
		for {
			err = sem.Acquire(context.Background(), 1)
			check(err)
			go func() {
				defer func() {
					bar.add(1)
					sem.Release(1)
				}()
				_, err = file.Read(b4)
				if err == nil {
					data[a] = math.Float32frombits(binary.BigEndian.Uint32(b4))
					a++
				}
			}()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	case 14:
		// float64 needs 8 bytes
		b8 := make([]byte, 8)
		data := make([]float64, all)

		a := 0
		for {
			err = sem.Acquire(context.Background(), 1)
			check(err)
			go func() {
				defer func() {
					bar.add(1)
					sem.Release(1)
				}()
				_, err = file.Read(b8)
				if err == nil {
					data[a] = math.Float64frombits(binary.BigEndian.Uint64(b8))
					a++
				}
			}()
			if err != nil {
				if err == io.EOF {
					break
				} else {
					log.Fatalln(err)
				}
			}
		}
		return &IDX{
			data: data,
			dims: dims,
		}
	default:
		log.Fatalln("Unknown file data type ... are you sure this is an IDX file?")
	}
	return &IDX{}
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
*/
