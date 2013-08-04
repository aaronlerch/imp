package main

import (
	"fmt"
	"os"
	"math"
	"math/rand"
	"time"
	"image"
	_ "image/jpeg"
)

type Point struct {
	coords []uint8
	n int
}

type Cluster struct {
	points []Point
	center Point
	n int
}

func main() {
	rand.Seed(time.Now().Unix())

	fname := "lenna.jpg"
	if len(os.Args) > 1 {
		fname = os.Args[1]
	}

	fmt.Println("Processing file", fname)

	f, err := os.Open(fname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	start := time.Now()
	process(img)
	duration := time.Now().Sub(start)
	fmt.Println("Finished processing, duration", duration)
}

func random(min, max int) int {
    return rand.Intn(max - min) + min
}

func euclidean(p1, p2 Point) float64 {
	s := 0.0
	for i := 0; i < p1.n; i++ {
		diff := int(p1.coords[i]) - int(p2.coords[i])
		pow2 := math.Pow(float64(diff), 2)
		s += pow2
	}
	return math.Sqrt(s)
}

func calculate_center(points []Point, n int) Point {
	vals := make([]int, n)
	for i := 0; i < n; i++ {
		vals[i] = 0
	}

	for _, p := range points {
		for i := 0; i < n; i++ {
			vals[i] += int(p.coords[i])
		}
	}

	plen := len(points)
	center_coords := make([]uint8, n)
	for i := 0; i < n; i++ {
		center_coords[i] = uint8(vals[i] / plen)
	}
    
    return Point{center_coords, n}
}

func process(i image.Image) {
	k, min_diff := 3, 1.0

	bounds := i.Bounds()
	points := make([]Point, 0, bounds.Max.Y * bounds.Max.X)

	// Scan the rows
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := i.At(x,y).RGBA()
			points = append(points, Point{[]uint8{uint8(r), uint8(g), uint8(b)}, 3})
		}
	}

	// Calculate k-means
	// 1. Take 3 random samples of the colors to start with as clusters,
	//    ensuring that none of them are duplicates
	clusters := make([]Cluster, 0, k)
	seen := make([]int, 0, k)
	for len(clusters) < k {
		idx := random(0, len(points))
		found := false
		for s := 0; s < len(seen); s++ {
			if idx == seen[s] {
				found = true
				break
			}
		}

		if !found {
			p := points[idx]
			seen = append(seen, idx)
			clusters = append(clusters, Cluster{[]Point{p}, p, k})
		}
	}

	// Loop until each cluster's center isn't moving more than the acceptable distance
	iteration := 0
	for {
		//fmt.Println("Processing iteration", iteration)
		iteration += 1
		plists := make([][]Point, k)
		for i := 0; i < k; i++ {
			plists[i] = make([]Point, 0)
		}

		for _, p := range points {
			smallest_distance := math.MaxFloat64
			idx := 0
			for i := 0; i < k; i++ {
				distance := euclidean(p, clusters[i].center)
				if distance < smallest_distance {
					smallest_distance = distance
					idx = i
				}
			}

			//fmt.Printf("Point %v assigned to cluster %d\n", p, idx)
			plists[idx] = append(plists[idx], p)
		}

		diff := 0.0
		for i := 0; i < k; i++ {
			old_cluster := clusters[i]
			//fmt.Printf("Old cluster %d has %d points with center at %v\n", i, len(old_cluster.points), old_cluster.center)
			center := calculate_center(plists[i], old_cluster.n)
			//fmt.Printf("New center of %d points: %v\n", len(plists[i]), center)
			new_cluster := Cluster{plists[i], center, old_cluster.n}
			clusters[i] = new_cluster
			//fmt.Printf("New cluster %d has %d points with center at %v\n", i, len(clusters[i].points), clusters[i].center)
			cluster_center_diff := euclidean(old_cluster.center, new_cluster.center)
			//fmt.Printf("Cluster %d's center has moved %v\n", i, cluster_center_diff)
			diff = math.Max(diff, cluster_center_diff)
		}

		//fmt.Printf("Comparing cluster center move diff max of %v against min_diff of %v\n", diff, min_diff)
		if diff < min_diff {
			break
		}
	}

	fmt.Printf("Completed in %d turns\n\n", iteration)
	for i := 0; i < k; i++ {
		c := clusters[i].center
		fmt.Printf("Color %d: %X%X%X\n", i, c.coords[0], c.coords[1], c.coords[2])
	}
}