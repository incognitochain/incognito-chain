package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/instruction"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

var samplePenalty = signaturecounter.Penalty{
	MinPercent:   50,
	Time:         0,
	ForceUnstake: true,
}

func Test_sortShardIDByIncreaseOrder(t *testing.T) {
	type args struct {
		arr []int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "testcase 1",
			args: args{
				[]int{15, 15, 3, 30},
			},
			want: []byte{2, 0, 1, 3},
		},
		{
			name: "testcase 2",
			args: args{
				[]int{1, 15, 3, 30},
			},
			want: []byte{0, 2, 1, 3},
		},
		{
			name: "testcase 3",
			args: args{
				[]int{30, 15, 3, 30},
			},
			want: []byte{2, 1, 0, 3},
		},
		{
			name: "testcase 4",
			args: args{
				[]int{30, 15, 45, 20},
			},
			want: []byte{1, 3, 0, 2},
		},
		{
			name: "testcase 5",
			args: args{
				[]int{190, 542, 208, 18, 674, 817, 808, 112},
			},
			want: []byte{3, 7, 0, 2, 1, 4, 6, 5},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sortShardIDByIncreaseOrder(tt.args.arr); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("sortShardIDByIncreaseOrder() = %v, want %v", got, tt.want)
			}
		})
	}
}

// [19,54,20,2,67,81,80,11]
func Test_calculateCandidatePosition(t *testing.T) {
	type args struct {
		candidate string
		rand      int64
		total     int
	}
	tests := []struct {
		name    string
		args    args
		wantPos int
	}{
		{
			name: "testcase 1",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGk4KtMcHSGEy6q7Ad5NPjGKakZoNowXd5xokQ3GYNSqLkmkicDFMRzHYk2qvUKfs6PbHFrLjnQNwQwX9inAzeBZdDeyRDNrPymyAwYkb5UDvvsqhx9fWF7Bm3TBYsZ5fKGLe9c5sok2HgKfZ8MUHyxXvYsmoAa4gPwECUULHXDFkh85XtMxEavYda1PMZCXr9fg9e6jV68RaRrNmodnJ77L7zcE9Dev6YAwPpSe3RpfmQ8Dj4tzhuiRuZiD4h1VEkDmhbuExWruL6VTaNpxRBkAhXgiktUS91WcXNq9CQPe793mvxedpJbyLsU5YsCoqw3bch5TUEoR1p9xD7fbzF6PmG868Cx9CJD73R2XFqFvijsLUnpoTVrZPfG9D6jVpCd1AxDGJv74FCWPQhm6xD7sUaRmpD",
				rand:      10000,
				total:     333,
			},
			wantPos: 39,
		},
		{
			name: "testcase 2",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGk4vSqVyeGkELuK9Zz94N2CjypNGdtskQKMFsseWJv377rYY5NGtTqnPq1PaNkGygmPeXdrjLjmCMcRJCWDRe7ie28Y6nm69a96d9JnDDYCvLsUbMMjmfgaPrZMaG1YmruauaeVqAhW8ahrubCtdAGJM9Bbb4yE4Lh4NBmWmQmsFSDJVQDmBTNE6M7ZuvcgQB3o3cMXGPuFpb7CpWBHvvjG15scXDLckgkjmzCgP8DGr72Y82uxeL2YULpToyjQuijYmY1sdHaAT5jdq8wuADYsMg5AthVpZRwNkdECtpVFey55VsG9mG283RwpQMyebqWASJvJPjpwQjTLQJrMcPdYjbyj6UFZFLoHBbj16A2a8awfEeqegR7TzUxMfnPNsBBfBTjEXZG56GFYLpzM1b885D4kkw",
				rand:      10000,
				total:     333,
			},
			wantPos: 52,
		},
		{
			name: "testcase 3",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGkQYyq7HFXd2y35K4p35YyNgoXZENZrgnhVKxLAzznRbDkLFvB6JAHPGg5vNfvesxU8xdKbDvi6ptt2UhU6BbvMEyuDg7ntDsH23pzs6cLbZALgfSFayfF6KvounPNRMkJ2piWYd8k7oXgC5sMhC4PcB6QxxKCW1een7KZKVHNpQQooCVUkTNiSuy25boa2Q3qrnEL6R9MWBykcm6ET14C7JyrrKXEz7oVaub2H3M8ByKCaib4ccGT9PungUPD8hvywNcKYYLXgM1SSt9kZKdhDBttUKa4X4PkmM2Ew9bzZLHQyESciGtgVv3yWsyQeGHMa7zcSrbFNYXdh5GRdqNJ4JwwpZyJvErYT5hy6HCDdBtjzVWLqFWQspGK72nhoJvfGjPMXYLYcvxvhMWA1uc36M4DbDD",
				rand:      10000,
				total:     333,
			},
			wantPos: 116,
		},
		{
			name: "testcase 4",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGk6wTR2KdLEEADpvc7pzYoxaqk4FMh2iHJRopTUaPwhMhGmndM92rZBNtXQfB2PeRo1YSGKHxRByvH92jF6FPqZLYaJPRoWJitWRsf7r3ReXf3Qmb7WwiFpSbNDQvVWy2HqqdVLhEaYaRPbqzVjLJ5c7fcc32FaPcknX1BUz7L6nWz7enaEMQPzt5LxCdu4NkHWwoNBADF938v8S9dztELMoPQiaruboEsiegjVL1PK9iQqpkT68RWJofnzqAS9nGkwn8jHmf2aDbxHJkYwWAsRfXTtTuJMK6K7kNakgssTNUzTKWs2sviYp1tkseUT26kbj4BXi2icnajCeMJDXMKD5YufvHsgJ4pZvNJvDobi7YQf8iweCCSwkABRqDQPY9qqmPzWugc2jLkEAc5iVcdHbpxZQU",
				rand:      10000,
				total:     333,
			},
			wantPos: 185,
		},
		{
			name: "testcase 5",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGkPeatBbQ3kUFg2EqpopZ1MkV3nTUz3KiNTQqaqZKMGn28FmxUYNJ3KFQAA2ocg3psnREazqFGcRQGd8pR7HhvYXrwqcx8BABj2WA7Qyqj5DLGj419GR2GLjueTsTKgtve3voRyj9EhErBCHKZNc1VVtFmfqzGk6kMKwLCcjv3yuSaxdx9k9odgyYcoiAcJwzWanj8r2oPKJK5FDNjLuQ8xRF8gktucz5VB84iTY9DRZ9ua8Wn6RRDR9U5i9gg69Wc4g5pZPv7mc7PZZGLakf941HB4FMKxQqiJLR6imZHyLhHWnMsN17aA5T7JmxH8UdeZXLNj73Komy8pQCGzfKXkn8uVwhwvwvxAxczShKKEAACfMtEAnfsirv6Gi2VL9AtFYq5Jx1vsfB3HBpAxp9xFjV8oHG",
				rand:      10000,
				total:     333,
			},
			wantPos: 114,
		},
		{
			name: "testcase 6",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGkPyc9JTWSLSmivsQGCgeD8vxTbTegwvLCREXrsywGwsgVMqtdYxmsknXmiAw16TAZhRsJ4DXrFiPhjVkt73VvjK1Q1cjcxjA2BkW4NHtAYSeBVkcUuk5einnjbevayfMEQ8WdGZfKMutVA5AMEammuUhC8BybH7o7BnWg43JqmqvaQXAXuFbYTbK1WCVuE9Lpgddv5dv6hpz7Yp8AGp3v2yn1PTrwFDxWvLfD7sL7qj42c7iZq4gZkcbf5CgyJ438eZnbf6g9vUCnKJLhMx9dhbZhZnAV1cbbo7BEJySw2kEQcVma5gnoYBbKtoJ5xRDQRZTwMk3g1a5eJ2u69Ripmv5vA1Cpt1Q9emQiDaw1VMVXHSbiYgEgCcNtZcsmxqYYYFGL8ZLZjL9tck4N4LFziGa6oEB",
				rand:      10000,
				total:     333,
			},
			wantPos: 322,
		},
		{
			name: "testcase 7",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGkBN1bTLWLWx35tLscRLVb77bFkeSLx911CiKs39cR9pA9YsDjrFEbRys9bNEFY8TesFDX3W89M5PuzyVwLgZm51KqSFpYxCXTnJnT9RkT5qr2KfkjbhgpfvvkLJV2YHwyPTbmKnbHcYXLLGzJeE8TogpZDDg38TckC3YR4xXezKaUR2thAfZDwnnSutrprKSkM6aDUP7SeYmqcEUYLN8HmF2wjcstPFfHu2hEY8PLYSbmMYbtPDp5sJnEQHHyfftRZJneaEJci9KiTuBNPfswj3LsKmDAmCZ5zqRkRpYjGyKYDhTWevyRvbf9tZskpfG7tR23VMYoLr5bEXwxUvdSsPpEWAs2xbHAazUk7MytBrVrgbRReeANFZdzRhPacNsgCRPvBzHAeL2eDMrfzH4XYqmfAha",
				rand:      10000,
				total:     333,
			},
			wantPos: 204,
		},
		{
			name: "testcase 8",
			args: args{
				candidate: "121VhftSAygpEJZ6i9jGkBMpYsJSyYtwUxuUPwfNBKqC44vmE4WsqRaJpvSFNZ6S2TDptppCLzZAc6zDxMBnaLaCxuraVhu1tAjqML9cgume5RmE1DviSeD8ZosA7e2Pomn1ijMexkqREiyjFZ6fcMJVafYHeLGM5nxpaJEhr4SRx78YKwxCBwSBTUFB5iE7fXxekhfQQTVgcNBeJE1Zjh7sVYkkS5FkKY5H8q4NHTVMf99DwnqCCFpURLr3qPyrwN3SPHkLV2AbVuA1PYsh2L3mZvmzSrm88phFYhTVgWdfAqwim7CuLx5shj4rvir1qFpqcyrEX3z4276k2XTjcJ1CQsv6vj8vHN4YLTGCpJx6ky2wk74rP32PKHwhQohnUwi6UAgmL1qmWDhpe6ZEjopdseLgheZnoQXLe9cwvtLHq55t",
				rand:      10000,
				total:     333,
			},
			wantPos: 276,
		},
		{
			name: "Temporary",
			args: args{
				candidate: key7,
				rand:      1250000,
				total:     265,
			},
			wantPos: 182,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPos := calculateCandidatePosition(tt.args.candidate, tt.args.rand, tt.args.total); gotPos != tt.wantPos {
				t.Errorf("calculateCandidatePosition() = %v, want %v", gotPos, tt.wantPos)
			}
		})
	}
}

func Test_assignShardCandidateV2(t *testing.T) {

	initTestParams()

	type args struct {
		candidates         []string
		numberOfValidators []int
		rand               int64
	}
	tests := []struct {
		name string
		args args
		want map[byte][]string
	}{
		{
			name: "testcase 1",
			args: args{
				candidates: []string{
					"121VhftSAygpEJZ6i9jGk4KtMcHSGEy6q7Ad5NPjGKakZoNowXd5xokQ3GYNSqLkmkicDFMRzHYk2qvUKfs6PbHFrLjnQNwQwX9inAzeBZdDeyRDNrPymyAwYkb5UDvvsqhx9fWF7Bm3TBYsZ5fKGLe9c5sok2HgKfZ8MUHyxXvYsmoAa4gPwECUULHXDFkh85XtMxEavYda1PMZCXr9fg9e6jV68RaRrNmodnJ77L7zcE9Dev6YAwPpSe3RpfmQ8Dj4tzhuiRuZiD4h1VEkDmhbuExWruL6VTaNpxRBkAhXgiktUS91WcXNq9CQPe793mvxedpJbyLsU5YsCoqw3bch5TUEoR1p9xD7fbzF6PmG868Cx9CJD73R2XFqFvijsLUnpoTVrZPfG9D6jVpCd1AxDGJv74FCWPQhm6xD7sUaRmpD",
					"121VhftSAygpEJZ6i9jGk4vSqVyeGkELuK9Zz94N2CjypNGdtskQKMFsseWJv377rYY5NGtTqnPq1PaNkGygmPeXdrjLjmCMcRJCWDRe7ie28Y6nm69a96d9JnDDYCvLsUbMMjmfgaPrZMaG1YmruauaeVqAhW8ahrubCtdAGJM9Bbb4yE4Lh4NBmWmQmsFSDJVQDmBTNE6M7ZuvcgQB3o3cMXGPuFpb7CpWBHvvjG15scXDLckgkjmzCgP8DGr72Y82uxeL2YULpToyjQuijYmY1sdHaAT5jdq8wuADYsMg5AthVpZRwNkdECtpVFey55VsG9mG283RwpQMyebqWASJvJPjpwQjTLQJrMcPdYjbyj6UFZFLoHBbj16A2a8awfEeqegR7TzUxMfnPNsBBfBTjEXZG56GFYLpzM1b885D4kkw",
					"121VhftSAygpEJZ6i9jGkQYyq7HFXd2y35K4p35YyNgoXZENZrgnhVKxLAzznRbDkLFvB6JAHPGg5vNfvesxU8xdKbDvi6ptt2UhU6BbvMEyuDg7ntDsH23pzs6cLbZALgfSFayfF6KvounPNRMkJ2piWYd8k7oXgC5sMhC4PcB6QxxKCW1een7KZKVHNpQQooCVUkTNiSuy25boa2Q3qrnEL6R9MWBykcm6ET14C7JyrrKXEz7oVaub2H3M8ByKCaib4ccGT9PungUPD8hvywNcKYYLXgM1SSt9kZKdhDBttUKa4X4PkmM2Ew9bzZLHQyESciGtgVv3yWsyQeGHMa7zcSrbFNYXdh5GRdqNJ4JwwpZyJvErYT5hy6HCDdBtjzVWLqFWQspGK72nhoJvfGjPMXYLYcvxvhMWA1uc36M4DbDD",
					"121VhftSAygpEJZ6i9jGk6wTR2KdLEEADpvc7pzYoxaqk4FMh2iHJRopTUaPwhMhGmndM92rZBNtXQfB2PeRo1YSGKHxRByvH92jF6FPqZLYaJPRoWJitWRsf7r3ReXf3Qmb7WwiFpSbNDQvVWy2HqqdVLhEaYaRPbqzVjLJ5c7fcc32FaPcknX1BUz7L6nWz7enaEMQPzt5LxCdu4NkHWwoNBADF938v8S9dztELMoPQiaruboEsiegjVL1PK9iQqpkT68RWJofnzqAS9nGkwn8jHmf2aDbxHJkYwWAsRfXTtTuJMK6K7kNakgssTNUzTKWs2sviYp1tkseUT26kbj4BXi2icnajCeMJDXMKD5YufvHsgJ4pZvNJvDobi7YQf8iweCCSwkABRqDQPY9qqmPzWugc2jLkEAc5iVcdHbpxZQU",
					"121VhftSAygpEJZ6i9jGkPeatBbQ3kUFg2EqpopZ1MkV3nTUz3KiNTQqaqZKMGn28FmxUYNJ3KFQAA2ocg3psnREazqFGcRQGd8pR7HhvYXrwqcx8BABj2WA7Qyqj5DLGj419GR2GLjueTsTKgtve3voRyj9EhErBCHKZNc1VVtFmfqzGk6kMKwLCcjv3yuSaxdx9k9odgyYcoiAcJwzWanj8r2oPKJK5FDNjLuQ8xRF8gktucz5VB84iTY9DRZ9ua8Wn6RRDR9U5i9gg69Wc4g5pZPv7mc7PZZGLakf941HB4FMKxQqiJLR6imZHyLhHWnMsN17aA5T7JmxH8UdeZXLNj73Komy8pQCGzfKXkn8uVwhwvwvxAxczShKKEAACfMtEAnfsirv6Gi2VL9AtFYq5Jx1vsfB3HBpAxp9xFjV8oHG",
					"121VhftSAygpEJZ6i9jGkPyc9JTWSLSmivsQGCgeD8vxTbTegwvLCREXrsywGwsgVMqtdYxmsknXmiAw16TAZhRsJ4DXrFiPhjVkt73VvjK1Q1cjcxjA2BkW4NHtAYSeBVkcUuk5einnjbevayfMEQ8WdGZfKMutVA5AMEammuUhC8BybH7o7BnWg43JqmqvaQXAXuFbYTbK1WCVuE9Lpgddv5dv6hpz7Yp8AGp3v2yn1PTrwFDxWvLfD7sL7qj42c7iZq4gZkcbf5CgyJ438eZnbf6g9vUCnKJLhMx9dhbZhZnAV1cbbo7BEJySw2kEQcVma5gnoYBbKtoJ5xRDQRZTwMk3g1a5eJ2u69Ripmv5vA1Cpt1Q9emQiDaw1VMVXHSbiYgEgCcNtZcsmxqYYYFGL8ZLZjL9tck4N4LFziGa6oEB",
					"121VhftSAygpEJZ6i9jGkBN1bTLWLWx35tLscRLVb77bFkeSLx911CiKs39cR9pA9YsDjrFEbRys9bNEFY8TesFDX3W89M5PuzyVwLgZm51KqSFpYxCXTnJnT9RkT5qr2KfkjbhgpfvvkLJV2YHwyPTbmKnbHcYXLLGzJeE8TogpZDDg38TckC3YR4xXezKaUR2thAfZDwnnSutrprKSkM6aDUP7SeYmqcEUYLN8HmF2wjcstPFfHu2hEY8PLYSbmMYbtPDp5sJnEQHHyfftRZJneaEJci9KiTuBNPfswj3LsKmDAmCZ5zqRkRpYjGyKYDhTWevyRvbf9tZskpfG7tR23VMYoLr5bEXwxUvdSsPpEWAs2xbHAazUk7MytBrVrgbRReeANFZdzRhPacNsgCRPvBzHAeL2eDMrfzH4XYqmfAha",
					"121VhftSAygpEJZ6i9jGkBMpYsJSyYtwUxuUPwfNBKqC44vmE4WsqRaJpvSFNZ6S2TDptppCLzZAc6zDxMBnaLaCxuraVhu1tAjqML9cgume5RmE1DviSeD8ZosA7e2Pomn1ijMexkqREiyjFZ6fcMJVafYHeLGM5nxpaJEhr4SRx78YKwxCBwSBTUFB5iE7fXxekhfQQTVgcNBeJE1Zjh7sVYkkS5FkKY5H8q4NHTVMf99DwnqCCFpURLr3qPyrwN3SPHkLV2AbVuA1PYsh2L3mZvmzSrm88phFYhTVgWdfAqwim7CuLx5shj4rvir1qFpqcyrEX3z4276k2XTjcJ1CQsv6vj8vHN4YLTGCpJx6ky2wk74rP32PKHwhQohnUwi6UAgmL1qmWDhpe6ZEjopdseLgheZnoQXLe9cwvtLHq55t",
				},
				numberOfValidators: []int{19, 54, 20, 2, 67, 81, 80, 11},
				rand:               10000,
			},
			want: map[byte][]string{
				0: {
					"121VhftSAygpEJZ6i9jGkQYyq7HFXd2y35K4p35YyNgoXZENZrgnhVKxLAzznRbDkLFvB6JAHPGg5vNfvesxU8xdKbDvi6ptt2UhU6BbvMEyuDg7ntDsH23pzs6cLbZALgfSFayfF6KvounPNRMkJ2piWYd8k7oXgC5sMhC4PcB6QxxKCW1een7KZKVHNpQQooCVUkTNiSuy25boa2Q3qrnEL6R9MWBykcm6ET14C7JyrrKXEz7oVaub2H3M8ByKCaib4ccGT9PungUPD8hvywNcKYYLXgM1SSt9kZKdhDBttUKa4X4PkmM2Ew9bzZLHQyESciGtgVv3yWsyQeGHMa7zcSrbFNYXdh5GRdqNJ4JwwpZyJvErYT5hy6HCDdBtjzVWLqFWQspGK72nhoJvfGjPMXYLYcvxvhMWA1uc36M4DbDD",
					"121VhftSAygpEJZ6i9jGkPeatBbQ3kUFg2EqpopZ1MkV3nTUz3KiNTQqaqZKMGn28FmxUYNJ3KFQAA2ocg3psnREazqFGcRQGd8pR7HhvYXrwqcx8BABj2WA7Qyqj5DLGj419GR2GLjueTsTKgtve3voRyj9EhErBCHKZNc1VVtFmfqzGk6kMKwLCcjv3yuSaxdx9k9odgyYcoiAcJwzWanj8r2oPKJK5FDNjLuQ8xRF8gktucz5VB84iTY9DRZ9ua8Wn6RRDR9U5i9gg69Wc4g5pZPv7mc7PZZGLakf941HB4FMKxQqiJLR6imZHyLhHWnMsN17aA5T7JmxH8UdeZXLNj73Komy8pQCGzfKXkn8uVwhwvwvxAxczShKKEAACfMtEAnfsirv6Gi2VL9AtFYq5Jx1vsfB3HBpAxp9xFjV8oHG",
				},
				2: {
					"121VhftSAygpEJZ6i9jGk4KtMcHSGEy6q7Ad5NPjGKakZoNowXd5xokQ3GYNSqLkmkicDFMRzHYk2qvUKfs6PbHFrLjnQNwQwX9inAzeBZdDeyRDNrPymyAwYkb5UDvvsqhx9fWF7Bm3TBYsZ5fKGLe9c5sok2HgKfZ8MUHyxXvYsmoAa4gPwECUULHXDFkh85XtMxEavYda1PMZCXr9fg9e6jV68RaRrNmodnJ77L7zcE9Dev6YAwPpSe3RpfmQ8Dj4tzhuiRuZiD4h1VEkDmhbuExWruL6VTaNpxRBkAhXgiktUS91WcXNq9CQPe793mvxedpJbyLsU5YsCoqw3bch5TUEoR1p9xD7fbzF6PmG868Cx9CJD73R2XFqFvijsLUnpoTVrZPfG9D6jVpCd1AxDGJv74FCWPQhm6xD7sUaRmpD",
					"121VhftSAygpEJZ6i9jGk4vSqVyeGkELuK9Zz94N2CjypNGdtskQKMFsseWJv377rYY5NGtTqnPq1PaNkGygmPeXdrjLjmCMcRJCWDRe7ie28Y6nm69a96d9JnDDYCvLsUbMMjmfgaPrZMaG1YmruauaeVqAhW8ahrubCtdAGJM9Bbb4yE4Lh4NBmWmQmsFSDJVQDmBTNE6M7ZuvcgQB3o3cMXGPuFpb7CpWBHvvjG15scXDLckgkjmzCgP8DGr72Y82uxeL2YULpToyjQuijYmY1sdHaAT5jdq8wuADYsMg5AthVpZRwNkdECtpVFey55VsG9mG283RwpQMyebqWASJvJPjpwQjTLQJrMcPdYjbyj6UFZFLoHBbj16A2a8awfEeqegR7TzUxMfnPNsBBfBTjEXZG56GFYLpzM1b885D4kkw",
				},
				3: {
					"121VhftSAygpEJZ6i9jGk6wTR2KdLEEADpvc7pzYoxaqk4FMh2iHJRopTUaPwhMhGmndM92rZBNtXQfB2PeRo1YSGKHxRByvH92jF6FPqZLYaJPRoWJitWRsf7r3ReXf3Qmb7WwiFpSbNDQvVWy2HqqdVLhEaYaRPbqzVjLJ5c7fcc32FaPcknX1BUz7L6nWz7enaEMQPzt5LxCdu4NkHWwoNBADF938v8S9dztELMoPQiaruboEsiegjVL1PK9iQqpkT68RWJofnzqAS9nGkwn8jHmf2aDbxHJkYwWAsRfXTtTuJMK6K7kNakgssTNUzTKWs2sviYp1tkseUT26kbj4BXi2icnajCeMJDXMKD5YufvHsgJ4pZvNJvDobi7YQf8iweCCSwkABRqDQPY9qqmPzWugc2jLkEAc5iVcdHbpxZQU",
					"121VhftSAygpEJZ6i9jGkBN1bTLWLWx35tLscRLVb77bFkeSLx911CiKs39cR9pA9YsDjrFEbRys9bNEFY8TesFDX3W89M5PuzyVwLgZm51KqSFpYxCXTnJnT9RkT5qr2KfkjbhgpfvvkLJV2YHwyPTbmKnbHcYXLLGzJeE8TogpZDDg38TckC3YR4xXezKaUR2thAfZDwnnSutrprKSkM6aDUP7SeYmqcEUYLN8HmF2wjcstPFfHu2hEY8PLYSbmMYbtPDp5sJnEQHHyfftRZJneaEJci9KiTuBNPfswj3LsKmDAmCZ5zqRkRpYjGyKYDhTWevyRvbf9tZskpfG7tR23VMYoLr5bEXwxUvdSsPpEWAs2xbHAazUk7MytBrVrgbRReeANFZdzRhPacNsgCRPvBzHAeL2eDMrfzH4XYqmfAha",
				},
				7: {
					"121VhftSAygpEJZ6i9jGkPyc9JTWSLSmivsQGCgeD8vxTbTegwvLCREXrsywGwsgVMqtdYxmsknXmiAw16TAZhRsJ4DXrFiPhjVkt73VvjK1Q1cjcxjA2BkW4NHtAYSeBVkcUuk5einnjbevayfMEQ8WdGZfKMutVA5AMEammuUhC8BybH7o7BnWg43JqmqvaQXAXuFbYTbK1WCVuE9Lpgddv5dv6hpz7Yp8AGp3v2yn1PTrwFDxWvLfD7sL7qj42c7iZq4gZkcbf5CgyJ438eZnbf6g9vUCnKJLhMx9dhbZhZnAV1cbbo7BEJySw2kEQcVma5gnoYBbKtoJ5xRDQRZTwMk3g1a5eJ2u69Ripmv5vA1Cpt1Q9emQiDaw1VMVXHSbiYgEgCcNtZcsmxqYYYFGL8ZLZjL9tck4N4LFziGa6oEB",
					"121VhftSAygpEJZ6i9jGkBMpYsJSyYtwUxuUPwfNBKqC44vmE4WsqRaJpvSFNZ6S2TDptppCLzZAc6zDxMBnaLaCxuraVhu1tAjqML9cgume5RmE1DviSeD8ZosA7e2Pomn1ijMexkqREiyjFZ6fcMJVafYHeLGM5nxpaJEhr4SRx78YKwxCBwSBTUFB5iE7fXxekhfQQTVgcNBeJE1Zjh7sVYkkS5FkKY5H8q4NHTVMf99DwnqCCFpURLr3qPyrwN3SPHkLV2AbVuA1PYsh2L3mZvmzSrm88phFYhTVgWdfAqwim7CuLx5shj4rvir1qFpqcyrEX3z4276k2XTjcJ1CQsv6vj8vHN4YLTGCpJx6ky2wk74rP32PKHwhQohnUwi6UAgmL1qmWDhpe6ZEjopdseLgheZnoQXLe9cwvtLHq55t",
				},
			},
		},
		{
			name: "8 Shards 8 Candidates Random Number: [500000 .. 1000000] Current Total Validators: [300 .. 400]",
			args: args{
				candidates: []string{
					key, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					19, 54, 20, 2, 67, 81, 80, 11,
				},
				rand: 800000,
			},
			want: map[byte][]string{
				0: {
					key, key5,
				},
				1: {
					key8,
				},
				2: {
					key4, key6,
				},
				3: {
					key3,
				},
				4: {
					key2,
				},
				7: {
					key7,
				},
			},
		},
		{
			name: "8 Shards 8 Candidates Random Number: [0 .. 500000] Current Total Validators: [300 .. 400]",
			args: args{
				candidates: []string{
					key, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					19, 54, 20, 2, 67, 81, 80, 11,
				},
				rand: 100000,
			},
			want: map[byte][]string{
				0: {
					key,
					key2,
					key8,
				},
				7: {
					key3, key4, key5, key6,
				},
				4: {
					key7,
				},
			},
		},
		{
			name: "8 Shards 8 Candidates Random Number: [1000000 .. 2000000] Current Total Validators: [300 .. 400]",
			args: args{
				candidates: []string{
					key, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					19, 54, 20, 2, 67, 81, 80, 11,
				},
				rand: 1250000,
			},
			want: map[byte][]string{
				0: {
					key4,
					key6,
					key8,
				},
				2: {
					key3,
				},
				3: {
					key,
				},
				7: {
					key2,
					key5, key7,
				},
			},
		},
		{
			name: "8 Shards 8 Candidates Random Number: [500000 .. 1000000] Current Total Validators: [200 .. 300]",
			args: args{
				candidates: []string{
					key, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					50, 33, 29, 47, 15, 2, 25, 64,
				},
				rand: 800000,
			},
			want: map[byte][]string{
				0: {
					key6,
				},
				1: {
					key2, key3, key4,
				},
				4: {
					key, key7,
				},
				5: {
					key5,
				},
				6: {
					key8,
				},
			},
		},
		{
			name: "8 Shards 8 Candidates Random Number: [0 .. 500000] Current Total Validators: [200 .. 300]",
			args: args{
				candidates: []string{
					key, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					50, 33, 29, 47, 15, 2, 25, 64,
				},
				rand: 100000,
			},
			want: map[byte][]string{
				0: {
					key4,
				},
				2: {
					key,
				},
				5: {
					key7, key8,
				},
				6: {
					key2, key3, key5, key6,
				},
			},
		},
		{
			name: "8 Shards 8 Candidates Random Number: [1000000 .. 2000000] Current Total Validators: [200 .. 300]",
			args: args{
				candidates: []string{
					key,
					key2,
					key3,
					key4,
					key5,
					key6,
					key7,
					key8,
				},
				numberOfValidators: []int{
					50, 33, 29, 47, 15, 2, 25, 64,
				},
				rand: 1250000,
			},
			want: map[byte][]string{
				3: {
					key7,
				},
				4: {
					key,
					key3,
					key5,
					key8,
				},
				5: {
					key2,
					key6,
				},
				6: {
					key4,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := assignShardCandidateV2(tt.args.candidates, tt.args.numberOfValidators, tt.args.rand)
			if len(got) != len(tt.want) {
				t.Errorf("assignShardCandidateV2() = %v, want %v", got, tt.want)
			}
			for k, gotV := range got {
				wantV, ok := tt.want[k]
				if !ok {
					t.Errorf("assignShardCandidateV2() = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(gotV, wantV) {
					t.Errorf("assignShardCandidateV2() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func Test_removeValidatorV2(t *testing.T) {
	type args struct {
		validators        []string
		removedValidators []string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Remove validators not found in list validators",
			args: args{
				validators:        []string{key},
				removedValidators: []string{key2},
			},
			wantErr: true,
			want:    []string{},
		},
		{
			name: "Found Validators In List Validators",
			args: args{
				validators:        []string{key},
				removedValidators: []string{key},
			},
			wantErr: false,
			want:    []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := removeValidatorV2(tt.args.validators, tt.args.removedValidators)
			if (err != nil) != tt.wantErr {
				t.Errorf("removeValidatorV2() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeValidatorV2() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_slashingSwapOut(t *testing.T) {

	initTestParams()
	initLog()

	type args struct {
		committees             []string
		substitutes            []string
		penalty                map[string]signaturecounter.Penalty
		minCommitteeSize       int
		numberOfFixedValidator int
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 []string
		want2 []string
	}{
		{
			name: "committee have min length",
			args: args{
				committees: []string{
					key, key0, key2, key3,
				},
				substitutes: []string{},
				penalty: map[string]signaturecounter.Penalty{
					key0: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key, key0, key2, key3,
			},
			want1: []string{},
			want2: []string{},
		},
		{
			name: "swap offset 3, one slash, spare one slash in fixed nodes, two normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				substitutes: []string{},
				penalty: map[string]signaturecounter.Penalty{
					key8: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key6, key7, key9,
			},
			want1: []string{
				key8,
			},
			want2: []string{
				key4, key5,
			},
		},
		{
			name: "swap offset 3, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8: samplePenalty,
					key6: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key5, key7, key9,
			},
			want1: []string{
				key6, key8,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 3, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key8: samplePenalty,
					key6: samplePenalty,
					key:  samplePenalty,
				},
				minCommitteeSize:       4,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key5, key7, key9,
			},
			want1: []string{
				key6, key8,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 2, one slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5,
				},
				substitutes: []string{
					key6, key7, key8, key9,
				},
				penalty: map[string]signaturecounter.Penalty{
					key:  samplePenalty,
					key8: samplePenalty,
					key5: samplePenalty,
				},
				minCommitteeSize:       5,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3,
			},
			want1: []string{
				key5,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 4 - 1, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12,
				},
				penalty: map[string]signaturecounter.Penalty{
					key:  samplePenalty,
					key8: samplePenalty,
					key5: samplePenalty,
				},
				minCommitteeSize:       10,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key6, key7, key9, key10, key11,
			},
			want1: []string{
				key5, key8,
			},
			want2: []string{
				key4,
			},
		},
		{
			name: "swap offset 4 - 0, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12, key13,
				},
				penalty: map[string]signaturecounter.Penalty{
					key:  samplePenalty,
					key8: samplePenalty,
					key5: samplePenalty,
				},
				minCommitteeSize:       10,
				numberOfFixedValidator: 4,
			},
			want: []string{
				key0, key, key2, key3, key7, key9, key10, key11,
			},
			want1: []string{
				key5, key8,
			},
			want2: []string{
				key4, key6,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := slashingSwapOut(tt.args.committees, tt.args.substitutes, tt.args.penalty, tt.args.minCommitteeSize, tt.args.numberOfFixedValidator)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapOut() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapOut() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_swapInAfterSwapOut(t *testing.T) {
	type args struct {
		committees       []string
		substitutes      []string
		maxCommitteeSize int
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 []string
		want2 []string
	}{
		{
			name: "push to max committee size",
			args: args{
				committees: []string{
					key0, key, key2, key3,
				},
				substitutes: []string{
					key4, key5, key6, key7, key8, key9,
				},
				maxCommitteeSize: 9,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8,
			},
			want1: []string{
				key9,
			},
			want2: []string{
				key4, key5, key6, key7, key8,
			},
		},
		{
			name: "push all substitute list but not max committee size",
			args: args{
				committees: []string{
					key0, key, key2, key3,
				},
				substitutes: []string{
					key4, key5, key6, key7, key8, key9,
				},
				maxCommitteeSize: 11,
			},
			want: []string{
				key0, key, key2, key3, key4, key5, key6, key7, key8, key9,
			},
			want1: []string{},
			want2: []string{
				key4, key5, key6, key7, key8, key9,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := swapInAfterSwapOut(tt.args.committees, tt.args.substitutes, tt.args.maxCommitteeSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("swapInAfterSwapOut() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("swapInAfterSwapOut() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("swapInAfterSwapOut() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func Test_createSwapShardInstructionV3(t *testing.T) {
	type args struct {
		shardID                byte
		substitutes            []string
		committees             []string
		minCommitteeSize       int
		maxCommitteeSize       int
		typeIns                int
		numberOfFixedValidator int
		penalty                map[string]signaturecounter.Penalty
	}
	tests := []struct {
		name  string
		args  args
		want  *instruction.SwapShardInstruction
		want1 []string
		want2 []string
		want3 []string
		want4 []string
	}{
		{
			name: "max committee size 8, one slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5,
				},
				substitutes: []string{
					key6, key7, key8, key9,
				},
				minCommitteeSize:       4,
				maxCommitteeSize:       8,
				typeIns:                instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidator: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key6, key7, key8, key9,
				},
				[]string{
					key5, key4,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{key0, key, key2, key3, key6, key7, key8, key9},
			want2: []string{},
			want3: []string{key5},
			want4: []string{key4},
		},
		{
			name: "max committee size 6, one slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5,
				},
				substitutes: []string{
					key6, key7, key8, key9,
				},
				minCommitteeSize:       4,
				maxCommitteeSize:       6,
				typeIns:                instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidator: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key6, key7,
				},
				[]string{
					key5, key4,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key6, key7,
			},
			want2: []string{
				key8, key9,
			},
			want3: []string{key5},
			want4: []string{key4},
		},
		{
			name: "max committee size 9, two slash, spare one slash in fixed nodes, no normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key8,
				},
				substitutes: []string{
					key6, key7, key9, key10, key11, key12,
				},
				minCommitteeSize:       4,
				maxCommitteeSize:       9,
				typeIns:                instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidator: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key6, key7, key9, key10,
				},
				[]string{
					key5, key8,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key4, key6, key7, key9, key10,
			},
			want2: []string{key11, key12},
			want3: []string{key5, key8},
			want4: []string{},
		},
		{
			name: "max committee size 12, swap offset 4 - 1, two slash, spare one slash in fixed nodes, one normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12,
				},
				minCommitteeSize:       10,
				maxCommitteeSize:       12,
				typeIns:                instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidator: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key12,
				},
				[]string{
					key5, key8, key4,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key6, key7, key9, key10, key11, key12,
			},
			want2: []string{},
			want3: []string{key5, key8},
			want4: []string{key4},
		},
		{
			name: "max committee size 12, swap offset 4, two slash, spare one slash in fixed nodes, two normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12, key13,
				},
				minCommitteeSize:       10,
				maxCommitteeSize:       12,
				typeIns:                instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidator: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key12, key13,
				},
				[]string{
					key5, key8, key4, key6,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key7, key9, key10, key11, key12, key13,
			},
			want2: []string{},
			want3: []string{key5, key8},
			want4: []string{key4, key6},
		},
		{
			name: "max committee size 12, swap offset 4 (push max), two slash, spare one slash in fixed nodes, two normal swap",
			args: args{
				shardID: 0,
				committees: []string{
					key0, key, key2, key3, key4, key5, key6, key7, key8, key9, key10, key11,
				},
				substitutes: []string{
					key12, key13, key14, key15, key16, key17,
				},
				minCommitteeSize:       10,
				maxCommitteeSize:       12,
				typeIns:                instruction.SWAP_BY_END_EPOCH,
				numberOfFixedValidator: 4,
				penalty: map[string]signaturecounter.Penalty{
					key5: samplePenalty,
					key:  samplePenalty,
					key8: samplePenalty,
				},
			},
			want: instruction.NewSwapShardInstructionWithValue(
				[]string{
					key12, key13, key14, key15,
				},
				[]string{
					key5, key8, key4, key6,
				},
				0,
				instruction.SWAP_BY_END_EPOCH,
			),
			want1: []string{
				key0, key, key2, key3, key7, key9, key10, key11, key12, key13, key14, key15,
			},
			want2: []string{key16, key17},
			want3: []string{key5, key8},
			want4: []string{key4, key6},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2, got3, got4 := createSwapShardInstructionV3(tt.args.shardID, tt.args.substitutes, tt.args.committees, tt.args.minCommitteeSize, tt.args.maxCommitteeSize, tt.args.typeIns, tt.args.numberOfFixedValidator, tt.args.penalty)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createSwapShardInstructionV3() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("createSwapShardInstructionV3() got1 = %v, want %v", got1, tt.want2)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("createSwapShardInstructionV3() got2 = %v, want %v", got2, tt.want2)
			}
			if !reflect.DeepEqual(got3, tt.want3) {
				t.Errorf("createSwapShardInstructionV3() got3 = %v, want %v", got3, tt.want3)
			}
			if !reflect.DeepEqual(got4, tt.want4) {
				t.Errorf("createSwapShardInstructionV3() got4 = %v, want %v", got4, tt.want4)
			}
		})
	}
}

func Test_getOrderedLowerSet(t *testing.T) {
	type args struct {
		mean               int
		numberOfValidators []int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		//TODO: Add test cases.
		// case: numberOfValidators > 0, mean == all shard (equal committee_size among all shard)
		// case: numberOfValidators > 0, mean > numberOfValidators of only 1 shard (in 8 shard)
		// case: only 1 shard
		// case: assign max half shard while possible more shard is belong to lower half
		// case: 8 shard, normal case, numberOfValidator are slightly different
		// case: 8 shard, 2 shard is much lower than other shard
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getOrderedLowerSet(tt.args.mean, tt.args.numberOfValidators); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getOrderedLowerSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssignRuleV3_Process(t *testing.T) {
	type args struct {
		candidates         []string
		numberOfValidators []int
		rand               int64
	}
	tests := []struct {
		name string
		args args
		want map[byte][]string
	}{
		//TODO: Add test cases.
		// case: numberOfValidators > 0, mean == all shard (equal committee_size among all shard)
		// case: numberOfValidators > 0, mean > numberOfValidators of only 1 shard (in 8 shard)
		// case: only 1 shard
		// case: assign max half shard while possible more shard is belong to lower half
		// case: 8 shard, normal case, numberOfValidator are slightly different
		// case: 8 shard, 2 shard is much lower than other shard
		{
			name: "case: 8 shard, 2 shard is much lower than other shard",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{440, 438, 442, 136, 89, 41, 309, 437},
				rand:               1000,
			},
			want: map[byte][]string{
				5: []string{
					"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",
				},
				4: []string{
					"121VhftSAygpEJZ6i9jGk9a5Kfw77TTKCB5FhKUBU1JJKrvogDS3g9JhfQXY4PP9xEzfnTRB43MhcpxmbR7qempyNsRu7k3oY59xphaP843bUnZWk17LiedGSaDUfc7xEB2jNt94rpm1FXF9hjPRijDqeqyyBhFV3uyqhmfCdnH1xxYpJW8XLk45Jhpf5vGZpy2qFn3vVUtmxk2TMdW68AsBvkT2PkFGQDqiWYRMBXStV1Npzxu3CUKamd74ZXA66tSY7rP1QE4vSFDCLX23rJE9tjVZuty74Edbin4ZzgG2PhLqvv7s3gpNzaY6oaaxRSbon7JWujnF3uv8o6DvdraGkWiJa1VhXrJjVDzNs5WVNtETGMPy58uUPkemv8oto6yCJzxiDgjUR9Yjy3mfnY3eNWV89BocWjNkBT4Y9JU8HMD8",
					"121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27",
					"121VhftSAygpEJZ6i9jGk9ZogHDZ369maVpypAooJmsxv3QK9apuYCZxe5iM61B1W8CgRRNuQuBEwBpjMM1bMBeL1buAa2LpokfP8FCKA2gfiLn9QCLew61SQYeTAgRG9rqouR5zph6ECxDbU7qYTytd5f8QFw2jedB9Y7C9eBCRgg6XHiKgaRK7WFN9iNYB3L4KGYGQCJZUTW7zLNxX2uQrpR3sPsERnsQYSfar2dLhH2d7R8Avywn1aZPH2hDLp1ytjv8kgbUM74fT7Dton9dCVm3AV4zgc39eX9j42kx2U7mXWeG7MKDZUBiAVAfWyxvecjpMt5juzuXvzRmiN5qkZJ8QVFAQ2XLSpszFZtuPJET5pSbfYn1StqvALbGKM5jU4n6oSGuQdFmSehMbkjZwZJ5oNBf1X77ax9hXigt3ejqh",
				},
				3: []string{
					"121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ",
					"121VhftSAygpEJZ6i9jGk9drax2iDha73FtSVnju8AYxEHLxLqrgcB5ocJPiJ3BBRcRgZ1TmTQxnEsSpSm3wEdaRd98Y7YEHBwrMsQdaPsA66MJeTxy9ZDpyAD82sWfYzHNA7Q8pjpBvCrvxKHQTz6NBRZXspvCtxozStN6mJMJWoMUyMBccZLgRMTN7dDXArcJVPtTVQWqjT15DToLbzY3qdnc1vdZDTq916qNdQ9PbCVwbswdqtdCxEwCoYo9uLS9gdkvJaJdU1wNuYFYvFgiAQFa6mgjNZWiDnLyYBtVX3VyfVGe4K8fRgG9bgj15ZG7UypBoQTjxxJJHDmMy23VHV3qSDr8bjLnhLVYgHmkpuHfhxFX2B9KXXhkc4XMgxxyC83HWaz2XvS1eNuTMVbKUd3tjCBkZQJszBDsKa5R7gJqH",
					"121VhftSAygpEJZ6i9jGk9a4wsKYVdTBhFsvZ4jwhzaWjFsondoUtyYpqF3hjWmPEUvMJF1Wh9NMhRVGaLYW8JtmS2JcP3Zq1L2AvYjYJaWm5CRqoYPz3DUEeJFSbUDdZPcp5YAL3FizX3J192zR3kWA967p6cZUgVP6e9LbHBdQBFGCQYRYKw8DNRCUzophTmoFeFwjWJF3EuXAWX2gv5ASj2Nem9YytNtjhZHCRS7Vz8pvgLdMJVa7B9fpS6oBV7nS4LWdeZc8NPC9VGKbTa1MCy7YkXfjbHfKpowEArnB9CLaLpWmSiaZruTiRxtZQZkU9z3YCMZW2dW5SHmwMGEseu6WwPPqgLz32tazKzNzHgiAJp561pxfCm7HF4r6VxtTqPKe8gjwLfDZqw3X2ew6rQ8Vo2csnjWrSQJAxYuMeinW",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AssignRuleV3{}
			if got := a.Process(tt.args.candidates, tt.args.numberOfValidators, tt.args.rand); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Process() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAssignRuleV3_SimulationBalanceNumberOfValidator(t *testing.T) {

	numberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}
	counter := 0
	isBalanced := false

	// assume no new candidates
	//1. assign new 40
	//2. swap out each shard 5 => 40
	// repeat until numberOfValidator among all shard is slightly different
	for {
		time.Sleep(10 * time.Millisecond)
		counter++
		randomNumber := rand.Int63()
		threshold := 10
		numberOfValidators, isBalanced = simulateAssignRuleV3(numberOfValidators, 5, randomNumber, threshold)
		if isBalanced {
			break
		}
	}
	t.Log(counter, numberOfValidators)
}

// Report:
// Threshold 10: roundly 40 times to balanced NumberOfValidators between shards
// Threshold 20: roundly 30 times to balanced NumberOfValidators between shards
// Threshold 20 -> 40: roundly 25 times to balanced NumberOfValidators between shards
func BenchmarkAssignRuleV3_SimulationBalanceNumberOfValidator(b *testing.B) {

	initialNumberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}

	// assume no new candidates
	//1. assign new 40
	//2. swap out each shard 5 => 40
	// repeat until numberOfValidator among all shard is slightly different
	counters := []int{}
	for i := 0; i < 1000; i++ {
		counter := 0
		isBalanced := false
		numberOfValidators := make([]int, len(initialNumberOfValidators))
		copy(numberOfValidators, initialNumberOfValidators)
		for {
			//time.Sleep(1 * time.Millisecond)
			counter++
			randomNumber := rand.Int63()
			threshold := 10
			numberOfValidators, isBalanced = simulateAssignRuleV3(numberOfValidators, 5, randomNumber, threshold)
			if isBalanced {
				break
			}
		}
		counters = append(counters, counter)
		b.Log(counter, numberOfValidators)
	}

	sum := 0
	for i := 0; i < len(counters); i++ {
		sum += counters[i]
	}
	b.Log(sum / len(counters))
}

func simulateAssignRuleV3(numberOfValidators []int, swapOffSet int, randomNumber int64, threshold int) ([]int, bool) {

	totalAssignBack := 0

	for i := 0; i < len(numberOfValidators); i++ {
		numberOfValidators[i] -= swapOffSet
		totalAssignBack += swapOffSet
	}

	candidates := []string{}
	for i := 0; i < totalAssignBack; i++ {
		candidate := fmt.Sprintf("%+v", rand.Uint64())
		candidates = append(candidates, candidate)
	}

	assignedCandidates := AssignRuleV3{}.Process(candidates, numberOfValidators, randomNumber)

	for shardID, newValidators := range assignedCandidates {
		numberOfValidators[int(shardID)] += len(newValidators)
	}

	maxDiff := calMaxDifferent(numberOfValidators)
	if maxDiff < threshold {
		return numberOfValidators, true
	}

	return numberOfValidators, false
}

func calMaxDifferent(numberOfValidators []int) int {
	arr := make([]int, len(numberOfValidators))
	copy(arr, numberOfValidators)
	sort.Slice(arr, func(i, j int) bool {
		return arr[i] < arr[j]
	})
	return arr[len(arr)-1] - arr[0]
}
