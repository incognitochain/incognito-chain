package committeestate

import (
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"testing"
	"time"
)

// [19,54,20,2,67,81,80,11]
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

func Test_AssignRuleV2_Process(t *testing.T) {

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
					"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",
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
					key1, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					19, 54, 20, 2, 67, 81, 80, 11,
				},
				rand: 800000,
			},
			want: map[byte][]string{
				0: {
					key1, key5,
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
					key1, key2, key3, key4, key5, key6, key7, key8,
				},
				numberOfValidators: []int{
					19, 54, 20, 2, 67, 81, 80, 11,
				},
				rand: 100000,
			},
			want: map[byte][]string{
				0: {
					key1,
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
					key1, key2, key3, key4, key5, key6, key7, key8,
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
					key1,
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
					key1, key2, key3, key4, key5, key6, key7, key8,
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
					key1, key7,
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
					key1, key2, key3, key4, key5, key6, key7, key8,
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
					key1,
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
					key1,
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
					key1,
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
			got := AssignRuleV2{}.Process(tt.args.candidates, tt.args.numberOfValidators, tt.args.rand)
			if len(got) != len(tt.want) {
				t.Errorf("Process() = %v, want %v", got, tt.want)
			}
			for k, gotV := range got {
				wantV, ok := tt.want[k]
				if !ok {
					t.Errorf("Process() = %v, want %v", got, tt.want)
				}
				if !reflect.DeepEqual(gotV, wantV) {
					t.Errorf("Process() = %v, want %v", got, tt.want)
				}
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
		{
			name: "numberOfValidators > 0, mean == all shard (equal committee_size among all shard)",
			args: args{
				mean:               10,
				numberOfValidators: []int{10, 10, 10, 10, 10, 10, 10, 10},
			},
			want: []int{0, 1, 2, 3},
		},
		{
			name: "numberOfValidators > 0, mean > numberOfValidators of only 1 shard (in 8 shard)",
			args: args{
				mean:               8,
				numberOfValidators: []int{1, 8, 8, 8, 9, 9, 10, 10},
			},
			want: []int{0},
		},
		{
			name: "only 1 shard",
			args: args{
				mean:               8,
				numberOfValidators: []int{8},
			},
			want: []int{0},
		},
		{
			name: "assign max half shard while possible more shard is belong to lower half",
			args: args{
				mean:               20,
				numberOfValidators: []int{1, 9, 9, 8, 9, 8, 10, 100},
			},
			want: []int{0, 3, 5, 1},
		},
		{
			name: "normal case, lower set < half of shard",
			args: args{
				mean:               12,
				numberOfValidators: []int{1, 9, 15, 12, 5, 17, 12, 20},
			},
			want: []int{0, 4, 1},
		},
		{
			name: "normal case, numberOfValidator are slightly different",
			args: args{
				mean:               15,
				numberOfValidators: []int{10, 9, 15, 12, 18, 17, 12, 20},
			},
			want: []int{1, 0, 3, 6},
		},
		{
			name: "normal case, 2 shard is much lower than other shard",
			args: args{
				mean:               292,
				numberOfValidators: []int{440, 438, 442, 136, 89, 41, 309, 437},
			},
			want: []int{5, 4, 3},
		},
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
		{
			name: "numberOfValidators > 0, mean == all shard (equal committee_size among all shard)",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{10, 10, 10, 10, 10, 10, 10, 10},
				rand:               1001,
			},
			want: map[byte][]string{
				0: []string{
					"121VhftSAygpEJZ6i9jGk9drax2iDha73FtSVnju8AYxEHLxLqrgcB5ocJPiJ3BBRcRgZ1TmTQxnEsSpSm3wEdaRd98Y7YEHBwrMsQdaPsA66MJeTxy9ZDpyAD82sWfYzHNA7Q8pjpBvCrvxKHQTz6NBRZXspvCtxozStN6mJMJWoMUyMBccZLgRMTN7dDXArcJVPtTVQWqjT15DToLbzY3qdnc1vdZDTq916qNdQ9PbCVwbswdqtdCxEwCoYo9uLS9gdkvJaJdU1wNuYFYvFgiAQFa6mgjNZWiDnLyYBtVX3VyfVGe4K8fRgG9bgj15ZG7UypBoQTjxxJJHDmMy23VHV3qSDr8bjLnhLVYgHmkpuHfhxFX2B9KXXhkc4XMgxxyC83HWaz2XvS1eNuTMVbKUd3tjCBkZQJszBDsKa5R7gJqH",
				},
				1: []string{
					"121VhftSAygpEJZ6i9jGk9dnzbXovyLBfFHs6cuMcjFEAB6FphEtuXPXCCNga16A1LkSHHSgLq9rFGQ496VnxHxDL8Ar6duhJxpEdcr75rkRFBzEvPNTkRYVbVcBUMtWd5PugsB6QjYt5CJQyUtzzbctC15AeX64aHrK1QwHJjFz8hnMz3eh8P8SPqKQ6zXhByqHm7YrcCY5uKZUn7CqM9RTwwJaqUhtqsHKyozBkfw26XhfkwvFh4vvL6McE5Ty1ztgiygUp8tt7haPEjVNGCqnwpPEB76oPVFTiKGePY1XqHy6aAFvuUBdmnNFEXrQPnh5xz5ULpq6PMJjVFvpu5kXcnaJbFBRZJcsJm1y69n56zUFKSg4LiNwqymf84U2SiPKT12cFehJedwfJkBMwnGpDAfaEZZvqcNRD9nVRUG4no9X",
					"121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ",
					"121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf",
				},
				2: []string{
					"121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27",
					"121VhftSAygpEJZ6i9jGk9dXXyF2XfP9G56ZoZ8hAtLMc3i33FWSqa3FJgMkc7pTs5qdaeJHAFfJpwjoGWazHWzknPSh8d319L9xMoz2TsLCefqPeP8Kqf24C8fuY9RCTCvnAmecnXL6SJyiVTP6Vjjhcvvdk8cQHVUSnnXxbuufgRckyw9Mc7VrgpG3qzBfeYCfWkurDmdVnyjh7jPsZQM1sBjRFiNSjQQo7HQgaMi2YQp9WGpE4kJfF265eqXySrT8BycKLnjunED9B1TU6WNs2e9aFB2u82tMzwoRwTHWbAgc9rixwM3UiAAnhiEhx56nkidGRsqo2LR5AQASHEUP6aHtnz1wcwovEPthXXChDhRuDAa5PHvfz57LQpWZANf9HV26J5wYFysJb4bN9vRcxTJdijRzDAt5QbBJs5DJo64N",
				},
				3: []string{
					"121VhftSAygpEJZ6i9jGk9a5Kfw77TTKCB5FhKUBU1JJKrvogDS3g9JhfQXY4PP9xEzfnTRB43MhcpxmbR7qempyNsRu7k3oY59xphaP843bUnZWk17LiedGSaDUfc7xEB2jNt94rpm1FXF9hjPRijDqeqyyBhFV3uyqhmfCdnH1xxYpJW8XLk45Jhpf5vGZpy2qFn3vVUtmxk2TMdW68AsBvkT2PkFGQDqiWYRMBXStV1Npzxu3CUKamd74ZXA66tSY7rP1QE4vSFDCLX23rJE9tjVZuty74Edbin4ZzgG2PhLqvv7s3gpNzaY6oaaxRSbon7JWujnF3uv8o6DvdraGkWiJa1VhXrJjVDzNs5WVNtETGMPy58uUPkemv8oto6yCJzxiDgjUR9Yjy3mfnY3eNWV89BocWjNkBT4Y9JU8HMD8",
					"121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwojMmVCgURwmS1oRqpfoFYhKwFUTzjyYduxBNegdgUw9VEYXJe3xSMBib24oBRhjbmK7LZ5qxRqKMsV7Lp49rPi2QGiizqiVKSsaAD2QMHEaC5ZBsGJvtRRMdGarHfe6yLtWAvcPLpsuSkYHfNni5Nh61LT1LwYLqSwDqNVKC3ew6RfRvmFfFTPhBGUcbhRKog6djipwKW9RZ7jvpgSiwifLriiz8h25ziR3Guh7k1cgyYm6TxQHrrLBRUgXKGvLkeAEi9ThrCsuQnbE61re6UmjwCcbPd9j53fzVmm4JJ19kMsyGUL2m7obSsdxvM",
					"121VhftSAygpEJZ6i9jGk9eKW2e2edvTfmZXDVdt8qTxwzuUVFHUtpf9JGKZN9damEuDoy381bwjwq6g5D4M6Zn2KUh2giSEktUca7nbvoM7L23aq9XXtsmyQKvVBseCVvUNmyERNHRZpzvzNCn6gwnzyMR58uBcibUBwV5441jYR18RxwwyKh2w8S6ogEmdrERAxdMYdxPXwj43Ve5aHnZtT8ZfV6vPPKnPmgyM95Bpw5ep1HmWoZvtF4s5WkqbCoaAYoBd94Bcysx4wzVQbvmU1SCD5hCpF3nDtE9n7G3SnNAerDE3DWiZt5GjpLtSpQyAjbaQekqpBnyKVqbsnPiDh5EynFqzETZLkt9p9hvAK2EuzQXBVZXpFwbwoiiw918LhZ2nw11xy2eQF6hGXG5GzaYYiBHmA4Fyd8rywPGA1QH4",
					"121VhftSAygpEJZ6i9jGk9ZogHDZ369maVpypAooJmsxv3QK9apuYCZxe5iM61B1W8CgRRNuQuBEwBpjMM1bMBeL1buAa2LpokfP8FCKA2gfiLn9QCLew61SQYeTAgRG9rqouR5zph6ECxDbU7qYTytd5f8QFw2jedB9Y7C9eBCRgg6XHiKgaRK7WFN9iNYB3L4KGYGQCJZUTW7zLNxX2uQrpR3sPsERnsQYSfar2dLhH2d7R8Avywn1aZPH2hDLp1ytjv8kgbUM74fT7Dton9dCVm3AV4zgc39eX9j42kx2U7mXWeG7MKDZUBiAVAfWyxvecjpMt5juzuXvzRmiN5qkZJ8QVFAQ2XLSpszFZtuPJET5pSbfYn1StqvALbGKM5jU4n6oSGuQdFmSehMbkjZwZJ5oNBf1X77ax9hXigt3ejqh",
					"121VhftSAygpEJZ6i9jGk9a4wsKYVdTBhFsvZ4jwhzaWjFsondoUtyYpqF3hjWmPEUvMJF1Wh9NMhRVGaLYW8JtmS2JcP3Zq1L2AvYjYJaWm5CRqoYPz3DUEeJFSbUDdZPcp5YAL3FizX3J192zR3kWA967p6cZUgVP6e9LbHBdQBFGCQYRYKw8DNRCUzophTmoFeFwjWJF3EuXAWX2gv5ASj2Nem9YytNtjhZHCRS7Vz8pvgLdMJVa7B9fpS6oBV7nS4LWdeZc8NPC9VGKbTa1MCy7YkXfjbHfKpowEArnB9CLaLpWmSiaZruTiRxtZQZkU9z3YCMZW2dW5SHmwMGEseu6WwPPqgLz32tazKzNzHgiAJp561pxfCm7HF4r6VxtTqPKe8gjwLfDZqw3X2ew6rQ8Vo2csnjWrSQJAxYuMeinW",
				},
			},
		},
		{
			name: "numberOfValidators > 0, mean > numberOfValidators of only 1 shard (in 8 shard)",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{1, 8, 8, 8, 9, 9, 10, 10},
				rand:               1000,
			},
			want: map[byte][]string{
				0: candidates,
			},
		},
		{
			name: "only 1 shard",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{8},
				rand:               1000,
			},
			want: map[byte][]string{
				0: candidates,
			},
		},
		{
			name: "assign max half shard while possible more shard is belong to lower half",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{1, 9, 9, 8, 9, 8, 10, 100},
				rand:               1000,
			},
			want: map[byte][]string{
				0: []string{
					"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",
				},
				1: []string{
					"121VhftSAygpEJZ6i9jGk9a5Kfw77TTKCB5FhKUBU1JJKrvogDS3g9JhfQXY4PP9xEzfnTRB43MhcpxmbR7qempyNsRu7k3oY59xphaP843bUnZWk17LiedGSaDUfc7xEB2jNt94rpm1FXF9hjPRijDqeqyyBhFV3uyqhmfCdnH1xxYpJW8XLk45Jhpf5vGZpy2qFn3vVUtmxk2TMdW68AsBvkT2PkFGQDqiWYRMBXStV1Npzxu3CUKamd74ZXA66tSY7rP1QE4vSFDCLX23rJE9tjVZuty74Edbin4ZzgG2PhLqvv7s3gpNzaY6oaaxRSbon7JWujnF3uv8o6DvdraGkWiJa1VhXrJjVDzNs5WVNtETGMPy58uUPkemv8oto6yCJzxiDgjUR9Yjy3mfnY3eNWV89BocWjNkBT4Y9JU8HMD8",
					"121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwojMmVCgURwmS1oRqpfoFYhKwFUTzjyYduxBNegdgUw9VEYXJe3xSMBib24oBRhjbmK7LZ5qxRqKMsV7Lp49rPi2QGiizqiVKSsaAD2QMHEaC5ZBsGJvtRRMdGarHfe6yLtWAvcPLpsuSkYHfNni5Nh61LT1LwYLqSwDqNVKC3ew6RfRvmFfFTPhBGUcbhRKog6djipwKW9RZ7jvpgSiwifLriiz8h25ziR3Guh7k1cgyYm6TxQHrrLBRUgXKGvLkeAEi9ThrCsuQnbE61re6UmjwCcbPd9j53fzVmm4JJ19kMsyGUL2m7obSsdxvM",
					"121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27",
				},
				3: []string{
					"121VhftSAygpEJZ6i9jGk9drax2iDha73FtSVnju8AYxEHLxLqrgcB5ocJPiJ3BBRcRgZ1TmTQxnEsSpSm3wEdaRd98Y7YEHBwrMsQdaPsA66MJeTxy9ZDpyAD82sWfYzHNA7Q8pjpBvCrvxKHQTz6NBRZXspvCtxozStN6mJMJWoMUyMBccZLgRMTN7dDXArcJVPtTVQWqjT15DToLbzY3qdnc1vdZDTq916qNdQ9PbCVwbswdqtdCxEwCoYo9uLS9gdkvJaJdU1wNuYFYvFgiAQFa6mgjNZWiDnLyYBtVX3VyfVGe4K8fRgG9bgj15ZG7UypBoQTjxxJJHDmMy23VHV3qSDr8bjLnhLVYgHmkpuHfhxFX2B9KXXhkc4XMgxxyC83HWaz2XvS1eNuTMVbKUd3tjCBkZQJszBDsKa5R7gJqH",
					"121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf",
				},
				5: []string{
					"121VhftSAygpEJZ6i9jGk9ZogHDZ369maVpypAooJmsxv3QK9apuYCZxe5iM61B1W8CgRRNuQuBEwBpjMM1bMBeL1buAa2LpokfP8FCKA2gfiLn9QCLew61SQYeTAgRG9rqouR5zph6ECxDbU7qYTytd5f8QFw2jedB9Y7C9eBCRgg6XHiKgaRK7WFN9iNYB3L4KGYGQCJZUTW7zLNxX2uQrpR3sPsERnsQYSfar2dLhH2d7R8Avywn1aZPH2hDLp1ytjv8kgbUM74fT7Dton9dCVm3AV4zgc39eX9j42kx2U7mXWeG7MKDZUBiAVAfWyxvecjpMt5juzuXvzRmiN5qkZJ8QVFAQ2XLSpszFZtuPJET5pSbfYn1StqvALbGKM5jU4n6oSGuQdFmSehMbkjZwZJ5oNBf1X77ax9hXigt3ejqh",
				},
			},
		},
		{
			name: "normal case, lower set < half of shard",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{1, 9, 15, 12, 5, 17, 12, 20},
				rand:               1001,
			},
			want: map[byte][]string{
				0: []string{
					"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",
				},
				4: []string{
					"121VhftSAygpEJZ6i9jGk9dPK5DE41KmqUhp2EFJXyyHNrQomEh7QW7icEh2zymM5953C2HN2kYYCbepkL37Qny6mgyxGiAxxqEEkCHnqK4FsNBDbGwojMmVCgURwmS1oRqpfoFYhKwFUTzjyYduxBNegdgUw9VEYXJe3xSMBib24oBRhjbmK7LZ5qxRqKMsV7Lp49rPi2QGiizqiVKSsaAD2QMHEaC5ZBsGJvtRRMdGarHfe6yLtWAvcPLpsuSkYHfNni5Nh61LT1LwYLqSwDqNVKC3ew6RfRvmFfFTPhBGUcbhRKog6djipwKW9RZ7jvpgSiwifLriiz8h25ziR3Guh7k1cgyYm6TxQHrrLBRUgXKGvLkeAEi9ThrCsuQnbE61re6UmjwCcbPd9j53fzVmm4JJ19kMsyGUL2m7obSsdxvM",
					"121VhftSAygpEJZ6i9jGk9drax2iDha73FtSVnju8AYxEHLxLqrgcB5ocJPiJ3BBRcRgZ1TmTQxnEsSpSm3wEdaRd98Y7YEHBwrMsQdaPsA66MJeTxy9ZDpyAD82sWfYzHNA7Q8pjpBvCrvxKHQTz6NBRZXspvCtxozStN6mJMJWoMUyMBccZLgRMTN7dDXArcJVPtTVQWqjT15DToLbzY3qdnc1vdZDTq916qNdQ9PbCVwbswdqtdCxEwCoYo9uLS9gdkvJaJdU1wNuYFYvFgiAQFa6mgjNZWiDnLyYBtVX3VyfVGe4K8fRgG9bgj15ZG7UypBoQTjxxJJHDmMy23VHV3qSDr8bjLnhLVYgHmkpuHfhxFX2B9KXXhkc4XMgxxyC83HWaz2XvS1eNuTMVbKUd3tjCBkZQJszBDsKa5R7gJqH",
					"121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27",
					"121VhftSAygpEJZ6i9jGk9dXXyF2XfP9G56ZoZ8hAtLMc3i33FWSqa3FJgMkc7pTs5qdaeJHAFfJpwjoGWazHWzknPSh8d319L9xMoz2TsLCefqPeP8Kqf24C8fuY9RCTCvnAmecnXL6SJyiVTP6Vjjhcvvdk8cQHVUSnnXxbuufgRckyw9Mc7VrgpG3qzBfeYCfWkurDmdVnyjh7jPsZQM1sBjRFiNSjQQo7HQgaMi2YQp9WGpE4kJfF265eqXySrT8BycKLnjunED9B1TU6WNs2e9aFB2u82tMzwoRwTHWbAgc9rixwM3UiAAnhiEhx56nkidGRsqo2LR5AQASHEUP6aHtnz1wcwovEPthXXChDhRuDAa5PHvfz57LQpWZANf9HV26J5wYFysJb4bN9vRcxTJdijRzDAt5QbBJs5DJo64N",
				},
				1: []string{
					"121VhftSAygpEJZ6i9jGk9dnzbXovyLBfFHs6cuMcjFEAB6FphEtuXPXCCNga16A1LkSHHSgLq9rFGQ496VnxHxDL8Ar6duhJxpEdcr75rkRFBzEvPNTkRYVbVcBUMtWd5PugsB6QjYt5CJQyUtzzbctC15AeX64aHrK1QwHJjFz8hnMz3eh8P8SPqKQ6zXhByqHm7YrcCY5uKZUn7CqM9RTwwJaqUhtqsHKyozBkfw26XhfkwvFh4vvL6McE5Ty1ztgiygUp8tt7haPEjVNGCqnwpPEB76oPVFTiKGePY1XqHy6aAFvuUBdmnNFEXrQPnh5xz5ULpq6PMJjVFvpu5kXcnaJbFBRZJcsJm1y69n56zUFKSg4LiNwqymf84U2SiPKT12cFehJedwfJkBMwnGpDAfaEZZvqcNRD9nVRUG4no9X",
				},
			},
		},
		{
			name: "normal case, numberOfValidator are slightly different",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{10, 9, 15, 12, 18, 17, 12, 20},
				rand:               1000,
			},
			want: map[byte][]string{
				0: []string{
					"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",
				},
				6: []string{
					"121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf",
				},
				3: []string{
					"121VhftSAygpEJZ6i9jGk9ZogHDZ369maVpypAooJmsxv3QK9apuYCZxe5iM61B1W8CgRRNuQuBEwBpjMM1bMBeL1buAa2LpokfP8FCKA2gfiLn9QCLew61SQYeTAgRG9rqouR5zph6ECxDbU7qYTytd5f8QFw2jedB9Y7C9eBCRgg6XHiKgaRK7WFN9iNYB3L4KGYGQCJZUTW7zLNxX2uQrpR3sPsERnsQYSfar2dLhH2d7R8Avywn1aZPH2hDLp1ytjv8kgbUM74fT7Dton9dCVm3AV4zgc39eX9j42kx2U7mXWeG7MKDZUBiAVAfWyxvecjpMt5juzuXvzRmiN5qkZJ8QVFAQ2XLSpszFZtuPJET5pSbfYn1StqvALbGKM5jU4n6oSGuQdFmSehMbkjZwZJ5oNBf1X77ax9hXigt3ejqh",
				},
				1: []string{
					"121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ",
					"121VhftSAygpEJZ6i9jGk9dTsGQVZNEnYLWhjXjomneDCT9XC6zGjeTW8ENTJVPooNnQBRuRdpwe9DoNVdLF4j3iy6Ld2p1eNLK8ek1bNSNrHFVjtsHaQVpcoHBy6nGghA9y7pr6ne6W7MRdAXLPHEDgpAodnRVxFSc2zyUbA48XKQVRtSVUneKjcdhDHP1hY9gC1EYLWtqn4weLzHKA84w77rHFhwwV2sbseHDqsmtQa4k5aiTCcXoUdmMpEuvadtSmKR33wciA3FNr5h125ce9ge1eSFPznvXNsCy7sA1rc4YJxipgwGoDDWSftUfnh5vY37WbwwNLhmMRvxtjHP7WBkSrLrKdamr8TdUu93cQyrykYRxbCben7pK1N75NknsSqLSihBnMTLo8gcFDGaDtQixPiJxMkefZ3qfHHwdvaL27",
					"121VhftSAygpEJZ6i9jGk9eKW2e2edvTfmZXDVdt8qTxwzuUVFHUtpf9JGKZN9damEuDoy381bwjwq6g5D4M6Zn2KUh2giSEktUca7nbvoM7L23aq9XXtsmyQKvVBseCVvUNmyERNHRZpzvzNCn6gwnzyMR58uBcibUBwV5441jYR18RxwwyKh2w8S6ogEmdrERAxdMYdxPXwj43Ve5aHnZtT8ZfV6vPPKnPmgyM95Bpw5ep1HmWoZvtF4s5WkqbCoaAYoBd94Bcysx4wzVQbvmU1SCD5hCpF3nDtE9n7G3SnNAerDE3DWiZt5GjpLtSpQyAjbaQekqpBnyKVqbsnPiDh5EynFqzETZLkt9p9hvAK2EuzQXBVZXpFwbwoiiw918LhZ2nw11xy2eQF6hGXG5GzaYYiBHmA4Fyd8rywPGA1QH4",
					"121VhftSAygpEJZ6i9jGk9a4wsKYVdTBhFsvZ4jwhzaWjFsondoUtyYpqF3hjWmPEUvMJF1Wh9NMhRVGaLYW8JtmS2JcP3Zq1L2AvYjYJaWm5CRqoYPz3DUEeJFSbUDdZPcp5YAL3FizX3J192zR3kWA967p6cZUgVP6e9LbHBdQBFGCQYRYKw8DNRCUzophTmoFeFwjWJF3EuXAWX2gv5ASj2Nem9YytNtjhZHCRS7Vz8pvgLdMJVa7B9fpS6oBV7nS4LWdeZc8NPC9VGKbTa1MCy7YkXfjbHfKpowEArnB9CLaLpWmSiaZruTiRxtZQZkU9z3YCMZW2dW5SHmwMGEseu6WwPPqgLz32tazKzNzHgiAJp561pxfCm7HF4r6VxtTqPKe8gjwLfDZqw3X2ew6rQ8Vo2csnjWrSQJAxYuMeinW",
				},
			},
		},
		{
			name: "case: 8 shard, 2 shard is much lower than other shard",
			args: args{
				candidates:         candidates,
				numberOfValidators: []int{440, 438, 442, 136, 89, 41, 309, 437},
				rand:               1000,
			},
			want: map[byte][]string{
				5: []string{
					"121VhftSAygpEJZ6i9jGk9a5Kfw77TTKCB5FhKUBU1JJKrvogDS3g9JhfQXY4PP9xEzfnTRB43MhcpxmbR7qempyNsRu7k3oY59xphaP843bUnZWk17LiedGSaDUfc7xEB2jNt94rpm1FXF9hjPRijDqeqyyBhFV3uyqhmfCdnH1xxYpJW8XLk45Jhpf5vGZpy2qFn3vVUtmxk2TMdW68AsBvkT2PkFGQDqiWYRMBXStV1Npzxu3CUKamd74ZXA66tSY7rP1QE4vSFDCLX23rJE9tjVZuty74Edbin4ZzgG2PhLqvv7s3gpNzaY6oaaxRSbon7JWujnF3uv8o6DvdraGkWiJa1VhXrJjVDzNs5WVNtETGMPy58uUPkemv8oto6yCJzxiDgjUR9Yjy3mfnY3eNWV89BocWjNkBT4Y9JU8HMD8",
					"121VhftSAygpEJZ6i9jGk9a5oQvgecAms7BtyyjGxpySigmDAdo8af26UKNmXYjNUhFVp4NN6RpRJFTGn57w7evPi3HkaF8ToXsCJ7ceaZ5p6hrCHLN6tKm5sjBEo3yusZZayurNMsrLGRhBE2i8Xhxkns8uN8c9WY5kcwSVsyD9f7399fMzRqtMB7TjyE4ad2KWDmteZZWuZB79jYB5wHWcxRyUxYaQ761gT8oMJ6FKr7wdPYAeuJ7Pai71Xi9YdPUDNQ1dZ7Uq1m7wxavKKax14Tuf9onu9oZDTeat7SNK1PDxjvf2uwkEmcHAp7qzp8c7igm8X6VjC6685gdThcdEHPxiwDsu3UxQyXK1fqwSHDHx3Ff7w5xKeDK8zJNghCLBbZ3HowQbT7hAKqu3N5puMKw5cjQJndA4trRw5yuzXxbf",
				},
				4: []string{
					"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",
				},
				3: []string{
					"121VhftSAygpEJZ6i9jGk9dnzbXovyLBfFHs6cuMcjFEAB6FphEtuXPXCCNga16A1LkSHHSgLq9rFGQ496VnxHxDL8Ar6duhJxpEdcr75rkRFBzEvPNTkRYVbVcBUMtWd5PugsB6QjYt5CJQyUtzzbctC15AeX64aHrK1QwHJjFz8hnMz3eh8P8SPqKQ6zXhByqHm7YrcCY5uKZUn7CqM9RTwwJaqUhtqsHKyozBkfw26XhfkwvFh4vvL6McE5Ty1ztgiygUp8tt7haPEjVNGCqnwpPEB76oPVFTiKGePY1XqHy6aAFvuUBdmnNFEXrQPnh5xz5ULpq6PMJjVFvpu5kXcnaJbFBRZJcsJm1y69n56zUFKSg4LiNwqymf84U2SiPKT12cFehJedwfJkBMwnGpDAfaEZZvqcNRD9nVRUG4no9X",
					"121VhftSAygpEJZ6i9jGk9a1qL1ZjK3QhzqpDT4uJop7tnEfahvCVRVbKPsuH94uGMqj1a1nLQSAcQypiUP3yc1s3t4jCRae8Kf6VXJRyHngt9X9iT2yJexKgRhnjzTfJYU4VkXV3w3xwiFCp5pGnDdocA8aq1SUCVMwfKAnhxjDHQDDWMMaxSLvjU56ZqaNZtsCQGTaySZLxc2RmCgMUeC5JQkF4t3P2NbTaZzz9JreLtZqSL3DVLacPwXy2enu3QMQYcPbZRVbz2Rxuf8e3hHYwjbAqvDJayCk9iasJnmZKP2gVRQhHUcX7cED4U22TMzi2rE6FoVMebThoDB2Dp21BW21qthS26Rkxe7UbxTesss9Xk1LSQsh8tRV1yxyGJ3DgtY2csBicRmT5PxQ3j9FwdSViKX82M1u9qv7ZcQSnvrJ",
					"121VhftSAygpEJZ6i9jGk9drax2iDha73FtSVnju8AYxEHLxLqrgcB5ocJPiJ3BBRcRgZ1TmTQxnEsSpSm3wEdaRd98Y7YEHBwrMsQdaPsA66MJeTxy9ZDpyAD82sWfYzHNA7Q8pjpBvCrvxKHQTz6NBRZXspvCtxozStN6mJMJWoMUyMBccZLgRMTN7dDXArcJVPtTVQWqjT15DToLbzY3qdnc1vdZDTq916qNdQ9PbCVwbswdqtdCxEwCoYo9uLS9gdkvJaJdU1wNuYFYvFgiAQFa6mgjNZWiDnLyYBtVX3VyfVGe4K8fRgG9bgj15ZG7UypBoQTjxxJJHDmMy23VHV3qSDr8bjLnhLVYgHmkpuHfhxFX2B9KXXhkc4XMgxxyC83HWaz2XvS1eNuTMVbKUd3tjCBkZQJszBDsKa5R7gJqH",
					"121VhftSAygpEJZ6i9jGk9eKW2e2edvTfmZXDVdt8qTxwzuUVFHUtpf9JGKZN9damEuDoy381bwjwq6g5D4M6Zn2KUh2giSEktUca7nbvoM7L23aq9XXtsmyQKvVBseCVvUNmyERNHRZpzvzNCn6gwnzyMR58uBcibUBwV5441jYR18RxwwyKh2w8S6ogEmdrERAxdMYdxPXwj43Ve5aHnZtT8ZfV6vPPKnPmgyM95Bpw5ep1HmWoZvtF4s5WkqbCoaAYoBd94Bcysx4wzVQbvmU1SCD5hCpF3nDtE9n7G3SnNAerDE3DWiZt5GjpLtSpQyAjbaQekqpBnyKVqbsnPiDh5EynFqzETZLkt9p9hvAK2EuzQXBVZXpFwbwoiiw918LhZ2nw11xy2eQF6hGXG5GzaYYiBHmA4Fyd8rywPGA1QH4",
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

func TestAssignRuleV3_Mainnet(t *testing.T) {
	initNumberOfValidators := []int{
		282,
		280,
		299,
		275,
		277,
		283,
		276,
		270,
	}
	numberOfValidators := []int{
		282,
		280,
		299,
		275,
		277,
		283,
		276,
		270,
	}
	randomNumber := int64(-3847674119873397474)

	candidates0 := []string{
		"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",
	}
	m0 := AssignRuleV3{}.Process(candidates0, initNumberOfValidators, randomNumber)
	t.Log(m0)
	for k, v := range m0 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates1 := []string{
		"121VhftSAygpEJZ6i9jGkMLXYwHZHgGVvdmVBL9fp8kuiCxQLQbfNSQWsSZqimQREqam81RPdYhBdRCfGcx7bwogYvEPEFYELGY3DSH8rT8pveicRLd938DQjzAJpNsceh5grjPrsxNUCAXFtZdErsrFcLkuu4EwRjz8UVBqYy1C7HNvnDhZKXBSJSecbPHeAGeVibkEQd8S8Hss5SMytZPhCurJHVAws9fsFm3m8eu48csgZy7xsktjgLKvac2Dx3JrKtstzBL7dA89vVRxLSLVF4Q1E9fowoSD8brjPoWyx13TYx5YeFXgUDWR6UvKEJesepHFK239nticR41hsNNjNN7zwdDsY8RMaau4UL4kd1pfQ68ezkxPCFoGq5d1ApEUpZKuNtSfVQjggUjEDLQru4nuPa7s2Vu2nTfXmDTRnJiX", "121VhftSAygpEJZ6i9jGkPwh5KvLHJxb7g8oNGpoAsgahhZqLmBucDuHqExPNTrwFgqXiyZgo7abCzmD7rYerWuYqYBkL5cjfEABJtAiuUEJheMSPMXn75gx22SJbrb4AdduSdkM1pZRzUKvZ2KokbY87nBTqbdZ3XVu3LRUK6LPAVUTG1HLeyE4oSdmAcbmGnTT8UUJ2araaNQvuYmmw4DXAC8KcmWz2YTX1CfbWww7UTXFvRgfY3Nbep8JaBrYEmsxEPqTugYD5sGNzXAwg3gmDMSxyUHPk5YzEQNQFFJt5S6b3oTdAVUHH6kS4CChGpP8k3K6WLTrVZ4bgbRcNx6XCNzgzEtS3DpMRU8tg3jGnuwGj1gofFJTdZq2XTTfnk7najVrkrSkLgXjVzANC39hdS8WCjB67TEt4ESiD1QTdeMd", "121VhftSAygpEJZ6i9jGkJw6uZXShbx3M91GZQZbem7EsQT6GB7hJHN9sZ4CWCSjXnq6VBE4ftMcDSXosq7hxx6FiVB1heu1cUhN3JME88HiTdEiaDbG4oAooU14wsGEVLipYVWUVuqfk1u9YkC7KG3LxFUmnwenc2imzAGd2E9BJ6UxReHoEjk85fm4QV6WeSYmFKWwMkNeL8NAqS5uMs2DJ2URQFSUNATtuCAyqF9qBsdKW84yJrNsX3ndWpAqGxzifggcZYNZCAc2piq8HjddnuACT57UiSf8y5PSVvAxaEqQGNowvDKVWACNfa4bTWv7qgB35RAXVaT9vFvDLziUX5rBHi2FZymjLVfDz4nvKuDK8oPqKCssghmBSy6yZ4SnChYuZXatiNNvp6wsjnVGXmdRaQNyDEZzsU9n2iSWmRWz", "121VhftSAygpEJZ6i9jGk7HxMNsQcZuMWRbxWpnhPkB2mPi9t1XMmsrwdn6s92tbwWyFFBC4Dsyd26TuLLWWkiz7LUg2Z8Rz5U3YupxkWZMicUjQn7V16EiRLRB4fB7eAb2aZR2XuJwUbqTvBgQg5QJMgJH1Mwjz38B6iTU2yJE3VQc9pGFRTbSw5jarn3X2TNYQyBqze6ceDzDfbnav9Un5qN7dgneWxrTH6Lj22exL8TgQGh8movAmnymL1H8RDd5tpFDAWWsiEUaPwyWQ9UQ65ZaeB8J23CKZ4V4Lef3Bia2AU7m2CDUo4Lg5Ntp52ts4UeLmzNXyc3YiYQihv5spgxXpQaBY9daDFyhnWdM62shRiy88QCyYLhGJRJyeyd2ifai4SWVbMSyBqrpVPUxWmfjQpyzEMJqQusRqLYFnP74L", "121VhftSAygpEJZ6i9jGkLkVoWbDxgeg6KF56y4uQn6cncmmDMJhRNMEfzdofkAEC5ggjAar3jsmKhFEEjfzZX1bgbkek24r1acVSZspn5zqnL7b5BAVocs8wFrtwBtQ1ynqdsmRGb6pNTGASLpB11VUJmCZoiukxL2ttKD3cX4W4vTaMcfWu5XsLMKVzTaKZNqSeHdLvxiBrZ6gj8ZqRyHj5rtugJf5uQF938Ku77qGtoUZ2YrrkjKgheGtTrj7qT1JzfnYRScjrqufwJgsU8sHpJMGSVj7uErZKckkNSUN95MW9HVRK4Wd26d1QMhRAnWBiJ3VL7RJgZUKkMVdosujs427kUN6NznNCfmRiovZjvNGWNYNa1H3WvnY9oGu7uiriMDyJSi74QMMTKM3n7ZAawaYPCkuyr5V9Zi4vSfGjLTQ",
	}
	m1 := AssignRuleV3{}.Process(candidates1, initNumberOfValidators, randomNumber)
	t.Log(m1)
	for k, v := range m1 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates2 := []string{
		"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",
	}
	m2 := AssignRuleV3{}.Process(candidates2, initNumberOfValidators, randomNumber)
	t.Log(m2)
	for k, v := range m2 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates3 := []string{
		"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",
	}
	m3 := AssignRuleV3{}.Process(candidates3, initNumberOfValidators, randomNumber)
	t.Log(m3)
	for k, v := range m3 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates4 := []string{
		"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",
	}
	m4 := AssignRuleV3{}.Process(candidates4, initNumberOfValidators, randomNumber)
	t.Log(m4)
	for k, v := range m4 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates5 := []string{
		"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",
	}
	m5 := AssignRuleV3{}.Process(candidates5, initNumberOfValidators, randomNumber)
	t.Log(m5)
	for k, v := range m5 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates6 := []string{
		"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",
	}
	m6 := AssignRuleV3{}.Process(candidates6, initNumberOfValidators, randomNumber)
	t.Log(m6)
	for k, v := range m6 {
		numberOfValidators[int(k)] += len(v)
	}
	t.Log(numberOfValidators[6] - initNumberOfValidators[6])

	candidates7 := []string{
		"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",
	}
	m7 := AssignRuleV3{}.Process(candidates7, initNumberOfValidators, randomNumber)
	t.Log(m7)
	for k, v := range m7 {
		numberOfValidators[int(k)] += len(v)
	}
	for k, v := range numberOfValidators {
		t.Log("shard ", k, v-initNumberOfValidators[k])
	}

}

// TestAssignRuleV3_SimulationBalanceNumberOfValidator case 1
// assume no new candidates
// 1. re-assign 40
// 2. swap out each shard 5 => 40
// repeat until numberOfValidator among all shard is slightly different
func TestAssignRuleV3_SimulationBalanceNumberOfValidator_1(t *testing.T) {

	numberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}
	counter := 0
	isBalanced := false

	for {
		time.Sleep(10 * time.Millisecond)
		counter++
		randomNumber := rand.Int63()
		threshold := 10
		numberOfValidators, isBalanced = simulateAssignRuleV3(
			numberOfValidators,
			5,
			0,
			40,
			randomNumber,
			threshold)
		if isBalanced {
			break
		}
	}
	t.Log(counter, numberOfValidators)
}

// TestAssignRuleV3_SimulationBalanceNumberOfValidator case 2
// 1. 5 new candidates
// 2. re-assign 40
// 3. swap out each shard 5 => 40
// repeat until numberOfValidator among all shard is slightly different6+/
func TestAssignRuleV3_SimulationBalanceNumberOfValidator_2(t *testing.T) {

	numberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}
	counter := 0
	isBalanced := false

	for {
		time.Sleep(10 * time.Millisecond)
		counter++
		randomNumber := rand.Int63()
		threshold := 10
		numberOfValidators, isBalanced = simulateAssignRuleV3(
			numberOfValidators,
			5,
			5,
			40,
			randomNumber,
			threshold)
		if isBalanced {
			break
		}
	}
	t.Log(counter, numberOfValidators)
}

// BenchmarkAssignRuleV3_SimulationBalanceNumberOfValidator case 1
// 1. NO new candidates
// 2. re-assign 40
// 3. swap out each shard 5 => 40
// repeat until numberOfValidator among all shard is slightly different
// Report:
// Threshold 10: roundly 40 times to balanced NumberOfValidators between shards
// Threshold 20: roundly 30 times to balanced NumberOfValidators between shards
// Threshold 20 -> 40: roundly 25 times to balanced NumberOfValidators between shards
func BenchmarkAssignRuleV3_SimulationBalanceNumberOfValidator_1(b *testing.B) {

	initialNumberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}
	counters := []int{}
	maxCounter := 0
	for i := 0; i < 1000; i++ {
		counter := 0
		isBalanced := false
		numberOfValidators := make([]int, len(initialNumberOfValidators))
		copy(numberOfValidators, initialNumberOfValidators)
		for {
			//time.Sleep(1 * time.Millisecond)
			counter++
			rand.Seed(time.Now().UnixNano())
			randomNumber := rand.Int63()
			threshold := 10
			numberOfValidators, isBalanced = simulateAssignRuleV3(numberOfValidators,
				5,
				0,
				40,
				randomNumber,
				threshold)
			if isBalanced {
				break
			}
		}
		counters = append(counters, counter)
		//b.Log(counter, numberOfValidators)
	}

	sum := 0
	for i := 0; i < len(counters); i++ {
		if counters[i] > maxCounter {
			maxCounter = counters[i]
		}
		sum += counters[i]
	}
	b.Log(sum/len(counters), maxCounter)
}

// BenchmarkAssignRuleV3_SimulationBalanceNumberOfValidator case 2
// 1. 5 new candidates
// 2. re-assign 40
// 3. swap out each shard 5 => 40
// repeat until numberOfValidator among all shard is slightly different
// Report:
// Threshold 10: roundly 40 times to balanced NumberOfValidators between shards
// Threshold 20: roundly 30 times to balanced NumberOfValidators between shards
// Threshold 20 -> 40: roundly 25 times to balanced NumberOfValidators between shards
func BenchmarkAssignRuleV3_SimulationBalanceNumberOfValidator_2(b *testing.B) {

	initialNumberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}
	counters := []int{}
	maxCounter := 0
	for i := 0; i < 100; i++ {
		counter := 0
		isBalanced := false
		numberOfValidators := make([]int, len(initialNumberOfValidators))
		copy(numberOfValidators, initialNumberOfValidators)
		for {
			//time.Sleep(1 * time.Millisecond)
			counter++
			rand.Seed(time.Now().UnixNano())
			randomNumber := rand.Int63()
			threshold := 10
			numberOfValidators, isBalanced = simulateAssignRuleV3(numberOfValidators,
				5,
				5,
				40,
				randomNumber, threshold)
			if isBalanced {
				break
			}
		}
		counters = append(counters, counter)
		//b.Log(counter, numberOfValidators)
	}

	sum := 0
	for i := 0; i < len(counters); i++ {
		if counters[i] > maxCounter {
			maxCounter = counters[i]
		}
		sum += counters[i]
	}
	b.Log(sum/len(counters), maxCounter)
}

func simulateAssignRuleV3(numberOfValidators []int, swapOffSet int, newCandidates int, totalAssignBack int, randomNumber int64, threshold int) ([]int, bool) {

	for i := 0; i < len(numberOfValidators); i++ {
		numberOfValidators[i] -= swapOffSet
	}

	candidates := []string{}
	for i := 0; i < totalAssignBack+newCandidates; i++ {
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

func BenchmarkAssignRuleV3_ProcessDistribution(b *testing.B) {
	genRandomString := func(strLen int) string {
		characters := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
		res := ""
		for i := 0; i < strLen; i++ {
			u := string(characters[rand.Int()%len(characters)])
			res = res + u
		}
		return res
	}
	for testTime := 0; testTime < 100; testTime++ {
		fmt.Println("Test Distribution ", testTime+1)
		initialNumberOfValidators := []int{472, 470, 474, 168, 121, 73, 309, 437}
		candidates := make(map[int][]string)
		for sid, v := range initialNumberOfValidators {
			for i := 0; i < v; i++ {
				candidates[sid] = append(candidates[sid], genRandomString(128))
			}
		}

		numberOfValidators := []int{}
		candidateAssignStat := make(map[string][8]int)
		assignTimeList := []int{}
		for { //loop until there is 1 candidate is assign 8000 times (expect ~1000 for each shard, and other candidate also are assigned ~ 8000 times)
			reach8000Times := false
			swapCandidate := []string{}
			numberOfValidators = []int{}
			for sid := 0; sid < len(candidates); sid++ {
				numberOfValidators = append(numberOfValidators, len(candidates[sid]))
			}
			for sid, candidate := range candidates {
				candidates[sid] = candidate[5:len(candidate)]
				swapCandidate = append(swapCandidate, candidate[:5]...)
			}
			rand.Seed(time.Now().UnixNano())
			randomNumber := rand.Int63()
			assignedCandidates := AssignRuleV3{}.Process(swapCandidate, numberOfValidators, randomNumber)
			for shardID, newValidators := range assignedCandidates {
				candidates[int(shardID)] = append(candidates[int(shardID)], newValidators...)

				for _, v := range newValidators {
					stat := candidateAssignStat[v]
					stat[shardID] = stat[shardID] + 1
					candidateAssignStat[v] = stat

					candidateTotalAssignTime := 0
					for _, sv := range stat {
						candidateTotalAssignTime += sv
					}
					if candidateTotalAssignTime > 8000 {
						reach8000Times = true
					}
				}
			}

			if reach8000Times {
				break
			}
		}

		//check our expectation
		for k, v := range candidateAssignStat {
			//check if each candidate, it is assign to shard ID uniformly
			if !isUniformDistribution(v[:], 0.2) { // allow diff of 20% from mean
				fmt.Printf("%v %v", k, v)
				b.FailNow()
			}

			candidateTotalAssignTime := 0
			for _, sv := range v {
				candidateTotalAssignTime += sv
			}
			assignTimeList = append(assignTimeList, candidateTotalAssignTime)
		}

		//check if all candidate has the same number of assign time (uniform distribution)
		if !isUniformDistribution(assignTimeList, 0.1) { // allow diff of 10% from mean
			fmt.Printf("diff: %v", calMaxDifferent(assignTimeList))
			b.FailNow()
		}
	}
}

func isUniformDistribution(arr []int, diffPercentage float64) bool {
	sum := 0
	for _, v := range arr {
		sum += v
	}
	mean := float64(sum) / float64(len(arr))
	allowDif := mean * diffPercentage
	//fmt.Println(mean, allowDif)
	for _, v := range arr {
		if math.Abs(mean-float64(v)) > allowDif {
			fmt.Println(mean, v, allowDif, math.Abs(mean-float64(v)))
			return false
		}
	}
	return true
}

func assignCandidate(lowerSet []int, randomPosition int, diff []int) byte {
	position := 0
	tempPosition := diff[0]
	for randomPosition >= tempPosition && position < len(diff)-1 {
		position++
		tempPosition += diff[position]
	}
	shardID := lowerSet[position]

	return byte(shardID)
}

func Test_assignCandidate(t *testing.T) {
	type args struct {
		lowerSet       []int
		randomPosition int
		diff           []int
	}
	tests := []struct {
		name string
		args args
		want byte
	}{
		{
			name: "assign at last position",
			args: args{
				lowerSet:       []int{1, 3, 0, 5},
				diff:           []int{300, 160, 89, 1},
				randomPosition: 549,
			},
			want: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := assignCandidate(tt.args.lowerSet, tt.args.randomPosition, tt.args.diff); got != tt.want {
				t.Errorf("assignCandidate() = %v, want %v", got, tt.want)
			}
		})
	}
}
