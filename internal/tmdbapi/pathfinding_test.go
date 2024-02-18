package tmdbapi

import (
	"testing"
	"time"
)

func TestParallelSearch(t *testing.T) {
	tests := map[int]struct{
		src string
		dest string
		expectedLength int
	}{
		0: {
			src: "Reservoir Dogs",
			dest: "Pulp Fiction",
			expectedLength: 1,
		},
		1: {
			src: "The City of Lost Children",
			dest: "Empire of the Sun",
			expectedLength: 2,
		},
		2: {
			src: "Midsommar",
			dest: "Gravity",
			expectedLength: 3,
		},
		3: {
			src: "The Descent",
			dest: "Prisoners",
			expectedLength: 2,
		},
		4: {
			src: "Fight Club",
			dest: "Rounders",
			expectedLength: 1,
		},
		5: {
			src: "Kickboxer",
			dest: "Dirty Rotten Scoundrels",
			expectedLength: 3,
		},
	}
	client := New("Bearer eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiI1Mzg4YzAwZmExNWRjYTc0YjU1YmM1MzA1MTViM2RjNiIsInN1YiI6IjY1YTBhNGEzZDIwN2YzMDEyOGU3NDI2YiIsInNjb3BlcyI6WyJhcGlfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.qbxIEsv2jty4BiZjDuh9MCZRrFc-XFrRdqq2G8JF4RY", time.Second * 5)

	for i := 0; i < 1; i++ {
		for _, test := range tests {
			path, err := GetPath2(&client, test.src, test.dest)
			if err != nil {
				t.Errorf("%s, for %s to %s", err.Error(), test.src, test.dest)
				continue
			}
			length := len(path) - 1
			if length != test.expectedLength {
				t.Errorf("length for %s to %s incorrect with: %v, wanted length %d",
					test.src, test.dest, path, test.expectedLength,
				)
			}
		}
	}
}
