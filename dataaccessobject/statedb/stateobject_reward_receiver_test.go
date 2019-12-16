package statedb_test

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

var (
	incognitoPublicKey = []string{
		"116MM3ii2VJigSizK4NzRFE9V3qPFKaeHdv2TZSss9RrqZxv3g",
		"11DgxrvZJrRfzLGA7cauH7q2chwEBXZSPTZg3RhxGyDjXuFKw2",
		"11H2zSkF6pVEBn42oMnHZskp86HRLufBbQj442sdmFm3u1sDrF",
		"11KyzW7ehjpNeGXpfFtFzJf5J4rgyHiwDTZsXv9B6MDMjCm4e",
		"11LHJznLa8NwScmtHi5GN8ZbyZuHFFLPq3M5UcuwLE7woVENgB",
		"121RdRKUGkFzmwJZszCQGaRfZXRE38jc4dibbF3e1XiE22thXqB",
		"121dxsK2J8YUKYiAfXbeyYe2dZcfpKG91VQzebNGz4wavfkBvh4",
		"121gzUmqhoQ2msU8Pz41j8RiVPFAKHaTbzhK6Tudx54umuTCSs4",
		"121kYkUqFDBU5caBpQYbougpZE471rfXp1JG5GgzyhMVWwcc9MH",
		"122EpgSNpZDDBKQDqcfszP7aBaBZrVLxwkuj2DWcJW2dvdqf5hV",
		"122N3e1ZHSEpA8uBofCv26g2ehLkoLqX3TDhFd1QHPwkZrxHP8s",
		"122NDmtXNf9G1fKg6exwHshSYqGVw37BqYqpvMVjjKvL7HSB6GS",
		"122aTEJhwZVHeTz8LPXy3U67Xb5Fqsz1P74Ep89LE58ZQoErZUi",
		"122k7fyxKo5nGu4d3qhaKS593UHyCLmXNHcdAduUu6EWLonzuJ2",
		"122njBEsurPx8Wq7kenWgkmnUGEkr4GRs8oGACqwRxfofyQ8yN7",
		"122oWA5XwGsc4PxcHqGPRdWu4pmPNBcR99ty1UN3XRpyGXdBeqQ",
		"1236UwHPsccSuYpxG2p68pYYEBDXomvqpQ5kXw4yEJBq3UQuq4u",
		"123WWWLefZF9jva4NMRzQymRZHo4ThjJpiofJoZPeGo8LsfcS9x",
		"123XPPs8riJmJ3Fgzn2mUXQsFi1nZ3sBJHJp5kYCZTMN4hA9Bio",
		"123pt85RBYx6nEKEfJ1XhRD3K3cT31Ygs47WUkUcrRRAxM2bEvw",
		"123yAnXSuphekALH2pQvYVVxhyDibp5w6TTbJJ4X37uy8RhskAZ",
		"1242ciXrKPQrLVsAn6puhg2p3ueG5QEdw5FEJHo4m4KscWXGQQo",
		"124ULqAP665rgpK73hiUxWNvxPABrRkcFx3tupcwMuY1gVbjyUr",
		"124bpWYHnLWkcFt3yhWjqJha6Ash1cAFkpZLxTECdRw8TjNDRXp",
		"124kR3CRT4Lfy43a78ECaupETGH5Y2gM9WyNmhAUqwU4vc1v75c",
		"124o1DwTv2yvZSMnDi51ACSbabBrUxLYwhkoW99TbLrWJt3noJK",
		"124siFNLdshTr3DggggNehvkNAjfcuXnRGT5XyV9GTcXXYdv1cu",
		"125NkBzVDSUxSp5jFSLb5hFG5oBAH8kL92TU3j56xtgQ9pYdNj1",
		"125UjtbvxvvghiB8g7ifg9cSBsUKNLDd9VDxXPKj57fpszfWsHV",
		"125fbh7Cseh4XSXJkKdcMUWP9PaZUc9zXHWow8btcFPZABT7Mjx",
		"1266Y1TuNTQgdkHct6wa9SH9cf5rig2B667nDDEeNtZAehWYrZ8",
		"126ak1ZeKcsNTBVE5MUAU1ArsvQCnn63uUepreJ7mUphM3qZ3vN",
		"126cC9hwFtpM1GCeCSNsDKy585Nsq5bCm9xzrdpbbm4696EAcZD",
		"126wQUH1GHcpcDyBGyTjkCLgXa8AT2n2AqppKP9SWt4opkopney",
		"126xw17FFuoYgck8KBp7GPSmoXSsyXxyMnbaTxcrpuJex9fk6VR",
		"1272dLx5n3N9NR1B6DcXQMUv5eeuu7oP7uVzZf4Sy8u5U43hF2h",
		"1276miiTpTZJhWXXKyE995ZqvDSyWzP1m2vAZ8ZF4EXq6sw1iZY",
		"127Ln8WaKEwQ15WxDvCb2hjZiA28EzZCa1QuUNwCmkoRNT9X6sH",
		"127NwkAqTDvKkc2SvX1csw9gnzwXwa9sD7wJJFNxknZANqrQEsz",
		"127U3AcScscho3bK2qf881uVyKhNdQZ6WDYTPMPpKDPBEv56ZDp",
		"127WBx98XJDzt3Kv3zyKayTXARGGjsugMxeUkdxRwpoKjYCBEXw",
		"127hNvEghAswaHAJWV6dZZ5JUbKGdMvGAG4ejUFcVWsp3hqGFYV",
		"127sP95Tks7fyt7XayjHG5gqsmHMNfQX5sew8QRRhHxWdRJkzg4",
		"128EAF8YAnVeME9YkTWEbRPsWWuQxqd8rRhVMRsfCVSomErr3wG",
		"128KVxYszyaS8NrXJKtKsmhgwdDesYxdCxXwU1SQGd3R6qMKMtG",
		"128LabtuKfDMyX7s6T4ErUJnFCMq6UN3dy3UufKNiYq34dac9fW",
		"128UjJXGsKUabwgqjcmfXaRFVMu71jb7yVWa24kLY1YSDXxszoW",
		"128jWAmotWLNP1JDHXohimWGyrQiAWLse8wQYhvcbKQXuZZ6ik7",
		"1295C2Y4NEvG6WzoEtC5XJcixqgo5Qm4BWY6Qv9i8D2V1QP8XyX",
		"129XbghMK3kxUCxxqVhpYXnsbPUXwvtHcFYXmXHPyhiqZeyVU9x",
		"129YuwqRKCtZMMrdnqy1X5aAV9tk2NoZogZV1C2Q4yFcT3nkySX",
		"129hQxESkzDDyrceF9zcmmdsSf3mcg2M28wCaJph1ebm1eL4kEb",
		"12ABDcXo4Zj1fq7zg5YYbveSfFNMjhVSRenYFHjogU76YnYTTts",
		"12ANwoiW58j5syAt3DJmkeTvHruvaNSqpKPJBiWyuzZFXqHHG7z",
		"12An6dacpbMSgqr2UQNW73Qz5ZXxqEf626UX1NPdaVBGmJtCAKg",
		"12BBe6j9bTrYcp2qyKZGjJuPBPPGqwPV6vqM2gLhFbznCy6n1Ld",
		"12BNWgpguMbWEQW4mZhMPYtm4KiawBEF1KUfX55PsieDfdSvXGj",
		"12BRhb6BqzbzkBtw2VCRpu6iEdDzXYNXQm395yFCySTtb8FopJ2",
		"12BdVdRDp1VkGqzcD3CqsJANanUFfqfsZCyrjoYeqADv8bWgv3w",
		"12Bip6RFwCUYgSMjhMDXZKvzPKawGZNswHVZBpS35AaiQhcY9BM",
		"12BwnzNGhwgtMGsyrpBxpqYA4DCWZoyKCVaURGSNcYxV5WrMFPd",
		"12CEJXecrsk2hNzFLjXYqxkE7P9ev1nPMctK7MafDDuiDa5S73y",
		"12CM4CmvRPGg6p6HR6NDvQYpbucdsk3eM2EeaXpiySMZzv1FFjQ",
		"12CP1EUirQzB1mKBVRRg3EEkVXMnjwoJXYu6CQaKCDNshgsTkwZ",
		"12CTLgHbSxRmtBqhk6j2E1EXxyiGkwWb8r5SbjgAW2gjFGJFLfU",
		"12CpE6q2n5zxwSr1dkUZvdKaiWeJsqZ9psixgn8NvbX5ypJFvTd",
		"12CqemGFiMAobfT4LmAGFcQ6qDQWU6EER5EdG8ewqia3CmBxZzw",
		"12CrEhGcZ4rpcDAbcW6hATwN67UAZqce8sDmPcWDYChyhyhKnYW",
		"12D1JiLmVGt85D9JxbKitaZKBbv5cNpXpE4BnhEvkaDMiajEDXX",
		"12DByVJjzXFHBee3Y7vfQkvJUGTejcN56iatdn7HQJpMEThyg8S",
		"12DZUwWMBEqzJH7UWFUd8945DFgZA7QXSU4txVVVjuexJVsJAeH",
		"12DhoH96MyMigzbNiMejydpvWRNTd8vBML93n3WfzSy7oEhkHjj",
		"12Dn4fh7EVzCVoCtnqaHczCbkBW9SSB4nYArwswWcGrxC4RgrWC",
		"12DnN1LnFma8pkiJDk8eaYFgv4uDiSsKoqfi3ufZLtKGbGSY8Nb",
		"12Duta66L7eSTyrkaqxZXpdevgaDwQiLRKYEcT1G6hM5sdqu1Df",
		"12DyYRGdmoujg4mcB6ESf6ioMTNCyDHqxHComrUACzKgJTtPDa1",
		"12EB5QE1ig23UENvFGi3xJVU7VKB4uJhq2vDvLVkkfdJy5SmeXT",
		"12ESxYKoFTfV4eoj6vgqt91RBF2Ruu9sHmN9KTxvRsqqeekL5es",
		"12F482nwYc39Mbp4RYQ74twGFQR3oMEsR8HkxXheTwx8SiGgWaT",
		"12FCT9Ks1Q94S1iPwUYfgfmrUifwk2TT33U5xPaVEgJpoStjkr3",
		"12FHD5peCB31bQBoafNdtuB9cwE2gPFPdWh4V3jzu2DbNuHM16H",
		"12FKD6QDxsRr3s3vs5sviTqQFXmFosJSvVuZvFivf3D4iSGuM7S",
		"12FNSBd85rSeVeKrJs54wW5CCmiLQo8Bn7vfHoAMtayHiak6ULP",
		"12Fao7QSWCQdNgjX81q7GY8H6LtEZ9BaKJrvxJkCV47rNKDJ5rv",
		"12FhiXjA6E2e1WYxCsfhCRTCRLW6XrY8SmnHahubt4VXqSYNHr6",
		"12GMErrKfheSVGa5qM9H6ResHrhZ3LvrKJwDctrf9ABHE3RJPPm",
		"12GXAdDvmfCJFCk9Wog4KcdxcgMKNNYRjXfpq3Czmvf8HHWqdRa",
		"12GgCcAATgMfwqMD6ckc6d3gJFKzi5wP7JWCLbu4Fd6Cp8MLYET",
		"12H8o4d1p1yHpPRSfRD35dxGpr6NArnDjNDzzrqM4bApqUddZR3",
		"12HCSbvLDP4uyTCVQGJdWZvzcQ7A6gMYAQL8ZPxfC41L6UGP9cj",
		"12HFJHdZBFGgyy2Dfs1z6vwDmAQmqkjvvkoocrksqHiyPjKUU43",
		"12HHLCJtWkqtQMTCBmrMviud86iher8LqENKdtg4SvnNUurDdkj",
		"12HQG2dGVNpKHfZeWqyQK5Utw9Wfyer7UEtChWTNDptBeeo8LfV",
		"12HVn8gzFDp2Z4fZTs9x1m9Gm98juiwigYkKfhzA2NZzYxqruDX",
		"12HxaMUzVrd72Y6c23JUJL4HSBLfEjzWC9JZqm751Uwp6bzLx5T",
		"12J3VfMXsaxW7LpqB7gLisvFwFUao37Gy7agcuNuLpvAja8hSYD",
		"12JQW9Y85Kxj4e1f9DieFmtdfFvVm6k63yhxr5n4kdTWfZ4RhSZ",
		"12JS4Qjjr4xStYW3hzB7QLjWYVu1gjJtHgmSinHVZ1RMmYSoX7E",
		"12JSiJEBYJRszLc36q7nz3Csh8pFA66bLBznLzBNBJRM5g7a87i",
		"12JavhPH5n3gS57gGikJqdf77Cjm19btTri4K3S632RjJLSDZjJ",
		"12KWcgkupLGHbuVx5NV9Av2DC1dYSe45Yct1CjAER5ajmb59rfV",
		"12KjuPtthb8JoVWQdA5NZzUMNDy6uKKhC3oGrZGF5aYv2VdaS8X",
		"12KxJ8YA39W4LUtGwKpYMaVUTbRdPjMed8QYsC9NbFemSqZ9gNJ",
		"12L45PBwnfAXR4R1hgNgvzyH5JSg91nqx9GGXETrgx6VbKzrCsk",
		"12LAYfJWESCpYeuA2RNeD7XwvuhtFQ7FiAajcGkbrKPmGc5UmUN",
		"12LTsTXNYSk8G2GXemBPQqFyEcqNHdeY3ZhBS2RA36xyE5hmL5P",
		"12LeHTp4BvREQPq2d1FcnkT11znyZmfbRSFPYzggUvtApLfUwmu",
		"12M73TtZZPCX2Vja7fNWP4oZhEmrSXWeEfiRMrKwSSh4gYZuBtM",
		"12MfhahYZbZ7hgNNXivQVJTu4QCN8Vg1hngHqdyx2phKtVyQvH8",
		"12Mp2mftBcV26XZnph2F2BAksiin2SHT1P3VNt9W8uT4rhN1aQA",
		"12N47bvXhcdCv9FsuMPYHsjmbPK7nq1xQovwjGKKpKtmyBuL3sv",
		"12Nv54o4F9feAPxHXkDf1kJauzghByjY8FLFLQTz5fjqP8aaPmL",
		"12NxyyJREyNuzxL6Q1u8JVNqPPHBhPrnQzLGey8N251JrcCNixq",
		"12P12PZwCMNwRPM3W7YM9avpVYZYbvzYzW2yRdtoLBx1AyATCAZ",
		"12P15ufNRZLH76h3vZzPrG39EDtfVzjBZW7DDwSUWgKN7KriANh",
		"12PCuYsQUmQ3ZRZMZ5hmL6qTpn9gmjQqBjnkQykdZLAv1jrH98m",
		"12PEGQxB8C7ELU1CFNp82aSKjKv3PJVHRk95mpLnfexq7kykzGm",
		"12PR2k39QyuSn1AJCNMHqX3J2ohTohfVyQGxjiSNBzuC5r8kMf7",
		"12PeWT5UZd1ftqpg2aYBtgEgk7hzMocQxLHaxSmC5AEuifnwLwV",
		"12Phpcuu2SjFCSREM94RSNE6WF6JFKbc4HM96K8tK4fF1v75cuR",
		"12Pzkg5DSHg2vttZ9ed4MYP7ang14kwFC4sUnNV5Eev4G1p8sVx",
		"12Q3GEdVZqNkDKdnU3vEHvdPNybkfp4EVpazVn3SyyHwNdtcp7X",
		"12Q4HSWG7dzLorKoqXWA2TJW8f4zUDPUn4N2wtaZPFu2JxtwHn",
		"12QLcX5FBTKN7LPp4wWiZeFKT7rPVa4vThi3Qqah7BtUWR3NpDH",
		"12QzhyWShmwjFEaskmDAjzqhTT8AJtFFYkesDXrVKsayRNWbqnH",
		"12R9akNkdHNNu9kF5zXSe76i3tjHUkwUJq8GPkXEHVxESC6A3Jc",
		"12RR4B3wriwrM8gRmSbPSm2PDgRUkpuGsZ2esKRdScgDB1j5QHR",
		"12RRMGAEXdtD796sbmyHwNv9oiPPfGYa1r6KssMqXyiUMfT6r86",
		"12RTHejUENdkcegKRg1Qnv35hLa6XtHfY3wmRfnPdxkQuWwN2kc",
		"12Rf57WXxRVY15FhuJFFL57bMhgSZvQWuVoc5kbhK5UFnU87avV",
		"12RtBtjFwEUiMcHnqgAHTPk5fhsXZ74N2M7iQHqkJRcmFWQZBB8",
		"12S2vRF4nQWA3UUukEeDfyRGmK2NMvvyfaRQe5b6ugBt3yZQrqs",
		"12SA88WKyWLqoQByQLSBaqMANft14NyoamgkM7RU5jr3HWCfe2r",
		"12SLwMFz5vWzZdY7vVWk9pWfgsQeMYJKN68dWRWAsDQA5H6rux1",
		"12SX5qRcXmS99q1gHnhjmpriupZ8gqGsxuX1iY2MoHWz86pJnSs",
		"12Svw8yVnn7yHfmdN7qbNdwksJPFX3vscdrtpjAg7rb1uQhmVRz",
		"12T7aK2nKWfzhX1qaTDAQCaAcGVjfkQrqJYbbkPkrKbBRC2kih9",
		"12TLiKb7LwZx13mPvJswNSWPZLiNEqNdW7qrfdLtbBUJJokrH2H",
		"12TUKof8xZsKoiStBv9hGH2thRMv6macdeKgMSF8CT3tAUneYwn",
		"12TXUTvSR2kEB1va2JjxqRozq6hHvHuTLZ3kHMdEncToyoobBsj",
		"12TcPGiCiPxGVNYsvM4MqG3Jqfmih6MvcGPrLrzZTxbDK8PhawT",
		"12U5xcP3mXZMbYvTHwgtcKfBLHvCHxfrGVi8FDoqC6bodiVHuGy",
		"12U74gXJWMkgpGE88F5neua3aFJ8Jwh2Q9znxbXcZ9qf5mhWmga",
		"12UJqf8C5Er2nxxES7wotZrgXH6UTM6FTqM865YcbDRENLok8wb",
		"12UPXzxUiR6LTugSZsynLYJvNKQsqwS6QAgUJSCSR3ccj2mrqrF",
		"12UkBTpH6Z4KyjkSrSycgsk59QiTgHRPa3jq66tj3KURMtxhLB6",
		"12UrQwBzThYWFHexZH1QZL1qzUzfDWPb1HJAL3P5fZkTLxjLGxK",
		"12V27nQr5ResgdvZqTZJQLByiwP69GoJc5ZfaAnfZ7fdhkYVdBP",
		"12V2G4698wQQ5ahfsVsXc3YuzrbSspdCH46ZCvJp17inFzfh8vP",
		"12V6nSFmjBzkJRuhuCE9E15Rai6EZjndnvJdNZkkw1o5XtsYhdB",
		"12VTpUUx1824udpr8n25gzbYD1XaDRngzCzVxksS9iaL2xYAWEZ",
		"12VUQiPzvrK4GxmRVE95yhY1uyvMW1efz4GhRbnDJYHYPquXNDk",
		"12W9LGz8VTkQ5vEkP9pM13n8GUduq4MzayYWBwShBkXb1pE4kCY",
		"12WFCrzXjruCmKCSYHLYxk9AH2xidZq4RhbR5JHThAjUqiw6w52",
		"12Wcnrjvn5f3UZBCVdNNGHzi16NCqXjWVaAUYFFAf59CY2ywRPY",
		"12WfFKGE4yGsZAZXDJn53CPtbAjsFJt6KpSnja47XeYn67q48BU",
		"12WgcHbKbmFDXfRxeu2PEkyd5pQp2LmQnTR5obJkc9h6G2H7gio",
		"12Wjfo8hUe9ht29dAxWSTS4HVEPtFLk5SWaTRpP643Xm7gpXP9H",
		"12WpsPgTBPRvb121DwWdrJGzohZvAy1Ng7p2DyRn9vBH4EM5rQx",
		"12X6zCPfwapbtccfNVQypXtHFscCpheq516qCeijAVqkbUsCyeS",
		"12XW4rCjiUWwiWPNNUTpSBBaUFVBijxudMqVL9vpFa4PLG4e3Qu",
		"12XbceKzgrWgnXPpZ9w5XM6eVzMsidnEPbDxznbEvqo8aKDbP4t",
		"12XddT6eUmz9aySjvNB8o5bfXoAH6eiP6WFTG8E9R9CGmPyFpoM",
		"12XfjYoNfxLQCJvaqaBwGGrr2oZzwKhMXM8ngDW7e3YFQxM2AEe",
		"12Xn5fr5JQpjBL3hqjwLSkmvNQqsfQWKtJBsisG9UJJjCVbzYP4",
		"12Xwc7MHef87uAKrK8jJ5ic9Ww5YYnfY2mfY7pXmyf3NaRHk2qJ",
		"12YCA7DXq9qJJceYPq2MT5UxZgwcxQhi97fKniSEWApqHYgLTmR",
		"12YWBJB4zMQUZVffQDwziqcCYRyV56RfB8UCTmrjfKm7BXymgXn",
		"12Z3o8QbviQgs8RnuRJ1g9Tt4zvF3XVQMS9SrBQuNJqtZAHnPKb",
		"12Z86JAJWJshhmPaxW1g4okWQgge6yDGei2wtqy7nZneDugiWEh",
		"12ZNbGJAK4YKgxZpqnKBs5GafZRbAc5vc2c6LqBGSU4VxPDigm6",
		"12ZTav4H4ts8QVu4ZBhNaNZ2kH48LfkGmGa7YNBV2FkoWT5GcTp",
		"12ZZ63UnVfbAQZjV9mDDTuskNeLnkEVMd53kppdzuSonQK28bEZ",
		"12ZbHAcSBugaCmk7iJr7Yn8KUb5FyBmAho3m8YbcfsLpY5PFbCZ",
		"12Ziy3KRjp7ANxAhzysLircfYYVsHH8SnJ44SN2aBwmcVSNmoRu",
		"12ZnifzjXAbotcvF5jboJgFkdLm3iBz8yzXGArr8CNcSUFUWYMJ",
		"12a97chqL5kZfgcNp8CsTZSTu42qVd3hk2NzpDgTQrTdDH8Tf2c",
		"12aExQgAdaaXqXH8hd1AJkZJQJGKLYMxQD497SEvBuhgqK6pkSu",
		"12aejSTzKURPvHVkKwKUM5kDpvsXPkhhnVQLAwWbCqXJwXRcF5t",
		"12asgSb1tcpd65gxykWtQbuHDqXN4uUb3BcAWpDXbRkE5yFW6Se",
		"12auD4A3pCD4o1qtX6wMFHfSjponRkHy3xBgFeejSjasBKqQrSh",
		"12azQxJmFgMxLiaZgmNgupGnvFV5bwRNvMmE5YaRtWpTP2AF9DR",
		"12bPyCp9kSPFxTNL6tM6VzpjHqT3ZrmPvUDcWH9wGTfbDWuAkF3",
		"12bXtgfgh2ckateixvMazd1xS7rb55B5pFzw4Vzbo2LJA55NjTT",
		"12bXvnQB9KU4Pmp1a7ouyyC8ht5udG7hLibswaVCyTesccwcxte",
		"12bvTTJAjLRESSHq4txwJUcCZTnjBsgXUutfNa3ZMRnuHju5R2L",
		"12c3GZwMziPwjSdFTfJ3i5eVEedzVcc9ANUXMVCZ3KCyQjbsPjZ",
		"12c5Ea6GZ48umVzqycYNfoMvYfQNKPv338dSWpmhgedHoxGYNiQ",
		"12c7UyGiA5TMcsAT4DToBZtqWnwLfj16WtW9W9UvoWXqZvPABNd",
		"12cKizPZgV1t3diGGSw1eBhpnjQwPKvqddzk6vHoG5DhFEr49CX",
		"12cLGdUXbk2cMswMhp938Nv13sxAVAXPVXMzrRFfTUos6HuVxSa",
		"12cb3YVdWNNV8MPNGEdaPekLQHGwUBYSm73A8EFWzmQCLmzdn8P",
		"12cbF1LpD91agaw18i64wvjtRRpp1NtsHzznA9294zcx79ZrQdt",
		"12cfrM2S1nj3qdDN7PHd6VkwQ1F8H9u13ALmnC7vFNJehwgUT6X",
		"12ckceWvJSR2LxRvFRoEB2ArhuSMsVCHKurszK5ZnUnz9iMEZdL",
		"12ctsYgkUjcMftwWeHMC15ADTER12N6C7u8Jz86AbdnwXK25HPA",
		"12dcPG5kFWg7MmUqQtg5CZt7sVtFzeUNf5CC1ebxUdbrEnVtPhX",
		"12dfVFswVaF3Ah7Vg4dG7YgH8VmsDg4xeS2SsMhYRZuFQQPUu3i",
		"12dvUeid5ojJtBC8h78fX5Zqsx1RMktWeA8cXrVke3ZcxLKJQj8",
		"12eKn82dYE5PNXUHhd4kjQz8MSSQn5FC77Q58ApqTQkJzzjjks1",
		"12ePYYGMKyLAPQcEjXC2kQZmwmRRuL8iEQRDMm2zgX3vr2Wu8Z4",
		"12etWpGDarPDcMJ41PDmPyEJWyJWXjCue5aUQhotf6dghNBNDeN",
		"12etq43pAw7KokBq5Qd7WLiPqsXfL3FhD3LwVRYVq9WuPWiQEp6",
		"12evutT1nzujoagEjiE1bbEkgQpf9i8Mr4znYGgR1eT2t57w5w9",
		"12f2j3sBrLyHf85WRJUGAiR55ANvXDsd7DkE3MoZD3MgVrLUjZz",
		"12f8CWK3YMVb8GJoZnHDs71CEVZjCVxuroZvKwEZUd33snTmrTj",
		"12fH8h9DNLbm6STtgKhGkqfLh1k4v9ZqQ3WrPvyCWhqMZDhigoM",
		"12fa4uWmh22PVyvVUqMeGjcE21y39ZYvBZdW8e8vjaQ1Sf8UYJG",
		"12fiNT7RBmdHXdP23vQQhYkwa7zH3y3Xp4tWEFM9qp4eVKjYt7S",
		"12g6tMFnMxyNjQKVZE34ZrLFUdjeKvcqMVXP3TsjiokrYGqBhX",
		"12go14vBsHtsg3GYSMTodx6rKqD3UctwmHXP3gXFgXZkEa7LinL",
		"12gv5JEcN8rLFaWJgwuq4hiV6Y45ZFxc6HvvhgiBiwvi4vNvrGE",
		"12hDL3enda2C8mwg12d682eosL7tenRaXpdRzH2FEwLiY2p1qE8",
		"12hFHwVyHKh4UioUWmvn3x3LQ7kAB44R9Bxbs5REfw9ZQY8aEEN",
		"12hMWGNGpv3vKxyEQBnKVTjMCWoaEiDHmTvfDPAmhkkg6wRE7cc",
		"12hN5myVxpSdY3YZgVmpXAup2ytZGVDh88L8Rqb4Yev24zxfMUc",
		"12hv1LfsAs79iNeVpq97o95b13BcqBFsfMiGfQvkEpJKbkfyez6",
		"12iDyyLMPok96JfpPv9SSeAWTgPwEYTUtUL4GAfjvM4CD97utEk",
		"12iEnQiF6dVPUWoTxfydjqcReVTHjmuFbKL5C7HE1sxBhNzTnwn",
		"12iLjLUeVvuX2ZTMaYU9qXpgou7VCZtDQCSqvG9aSRDTGuAc71L",
		"12ifez5Ju8gyHx1frCYzJoE2H3JvpniJux3QnbNJvbYLKzSD6oM",
		"12ijpZ2QTdmzv2skZ1D5RW2PqXnPRgKNcgnb2mRjpSZ2XyXc81s",
		"12io1pB4UVA1NJ89SDQUTxR48tLtdBhtSfJrbZpvWzgmKjbU7im",
		"12je1EQKkUnuGci48KZLf4FcwgAbycDGUB6msbMLatmDAh8P2d9",
		"12jgFsEHzjqhojD3r9dvkzTGZZoTseTFE3HpjQouWngahktHdSG",
		"12jnLdbdQfUmBqTcm1GQTg4J4iDDajMHoVX5MsK1o3Qwpd9WChC",
		"12juh4Ex4ugmfkq71dWjip1pGuHk47iHWBGEfPxg5XJeAtxoS6T",
		"12k29vfC3teGisy4mGNDmpL3RKnJn6vREbbDB5P32jybm9FgvDm",
		"12k5nMT66aYn487cwm36FTF1Qpoxb5Pjf19DEWqU4zbwAopPNfu",
		"12kK2C3QC4WmLHgx1XJQCqJHDSBzb1LPjbsFeF6gHaX83LEZLQE",
		"12kLh7rggqCuMpxaghrcXdADr7X8NVWj3Ys35x87PbVZRWRX8TD",
		"12kQJbCTofnB6xxmu293yeTqBqop8ZVHhqiQZZRjnNoeGqWpQce",
		"12kUFuTdJJgpSbhcWvcCnrvZsUnfa9gqn42Z8bWwtvvFMJZCfxb",
		"12m8NzJoocNgUd4zZLyXS9tYx1wzw8GByeGpfbmaocZo2gEKsxw",
		"12mrAKuNzm26rrd3quT4HXYb7bSwcfzdTQ5hg2W2hxphkq49onp",
		"12mrQkEVdBxUVT7b2afKggJkmTRQ6UAgtnJVthxy478bkPuLgzn",
		"12mwFDq4ww1enh7eUuQKr78RNw8dVYa4qMoaV1aqmSw23KaAeEb",
		"12n9G3qN4JBsfS24rHeSQSMW9xmJfvqkevMw1Y28HjtVFhRPgk8",
		"12nVyqCHSaPiq4ZeSAsBAZ3po2xcRrEvYC37K2S2RzZqYg3ho9N",
		"12o3vjds2ysZBrbmyAxV4Rkixiz3ibHVYC39SUN7TJXhHastMfz",
		"12o82LkqxDtU6vnjTZj5vCLoQ9yExqxADLWKHfHx6wWMsLZ7sC3",
		"12o9emgakf259RN5bVLQ4QZRz52ySWpUMG3A4Krwif3jh6n6XTH",
		"12oGVRBgoFNxFvUuCCvBm751YUNf12A1dSioEWghS2AE5nWpUne",
		"12oHS7kfD3f8gA7Js4EX4yntkQeahpxLvwBACKy81hqU5GYPZJ4",
		"12oiQRJYbBBkr6HQWAzMaGA7obTLgm92rswUsAxJV5diXRuA4AY",
		"12ouapmv5jA2g2BP3x7wETj6U72RmZLMox71e6tLvPJsACJLhEi",
		"12p1Ercdc41GLuiAk6nuP2y7JMMnrBo2g39XvnXwdDqoKDGwXQp",
		"12pH2dwoe8tu2QfEMqQznN5sS9mHStkHbwTBEH57Rc783vFkZDC",
		"12pNFs3vtYDBc142Mmn3SUjqvaV7fkfXfNJHb8VhEqhnkFVeFp3",
		"12pQYYu484HA18YgtKCT3SgbVtLMr5PFpR5uV17DYoM7hD9BaD9",
		"12pQwK6QgFFHbwjuvXMGeA46T4m6iRXXVGriJwiiurhBEsikab2",
		"12paMPJ2HrkjvNFcEavjkd25WTNPRdxUSVgSmMKRDRJnCREZLZq",
		"12ppJuoow9LKXc3iiTNV3pWPJ1M6pE5LRjLKAgXFcMKvzSuVXkT",
		"12pv88XyvJPmRsLDxaqV3Hwm9Qq8k1sYkH4Nx2JUYwD5hpMd8Jq",
		"12q7tH1VfDMMcDNMDoHAqJySJALwfRGiRjtYAj7dRxh9184TFHj",
		"12qLs45UxnWMKGptaRmE3y5LbYfxWF7gkYQqahKJXNYwbeAjZH9",
		"12qRW9mB3STWkHntQ15L3a4tsbx1ogxYaJJUSKyNoub37iJ5f7n",
		"12qdUhATi6xW8U64v864KK7ksapP5MTNXUHEGWLyshCvEBWmEVT",
		"12qmyrsaHVKQ5s2rZ9s6Mp1zv3AAFzGQGa4D8W3LCigYcsjVtiR",
		"12qsa7VrTLbPFfZ8xhfYSujdoNNfasMnVubvTpjKdh4w99quN3q",
		"12rN29Chq5gcozKBo8wvfJr8SYFjhHzhbaQsSMNa18xyytHy7pw",
		"12rYj5wrSpNVXCZ3oesSuqsGZa3oTcuxUHtG66Nx2PfMMih3u7q",
		"12rgotB8bcyz4NBnJAEtQidVHRpSR4HWHj5Dz6UKaDbgxApM1hB",
		"12rr8GXQKGRSb9fWSnt9hrqCGuz9QNtsRvvjPDCy5DCqfnuHtYB",
		"12sBzjw6k7SWRGypGgkJisTcUXPdorYQHthyFvvsy4R18n9Cg7z",
		"12sGUQ6qpBwe3g4bLAA9hiZZEJmzgZS1PhTKcv7T6UjxD3AnHxv",
		"12sJXV5kqVSvYv9tccPyKwZo8WUbZ9LRCfxXFZe56WLLh1Esd9x",
		"12sRD3bLB9K1m5rAoXKdTjMfX6akTBeKN2LtvR7s3EYYbASPGy1",
		"12sTmQx4zYRDwF1dJgnAVmMHodyK9uvEYG2FnCyAiSBKzHsp5iR",
		"12sbvqDiunKUxudBJnyr9Bge69v7mEU4A9X9LH4rDwpi5bBu4Xn",
		"12svt4q61KZCnA9N812whLxpTgAGZB6gT3PuFTU7XE9wE4VAX9b",
		"12tCL5ewATpcapuy8RYCc4MTug9kjF3Q7MbQzLQcDbvkHUhr7sj",
		"12tZGkNEJKcc9YSwchfUBw9auSB77USqTwRfWFnBxcz9cTKD2rP",
		"12tnDMirfrJzLeHgnsEeg9AmbyvGQNaHvEvF5VLEqtjXgURDjNj",
		"12traJnxJp25DGksQ2xAo5rxVo2LmXWmXBwQQCiQmpPLZizB5NX",
		"12tu2sgfmgSkYaFq2LRjpodkg7ZRWwfA9LXzeML2o8Gu6uwG5Jb",
		"12uBPEFCDeU6Zy9jFsaGPJ93KHPviKiq1zzsRsJ5X1fjasjYw8a",
		"12uGwjCunXpxJ2x1Nk6ogvUnhwnmUB7PZhNCbwhauSinp3Z9WkK",
		"12uHmQwW57dAKftmJjMQiL8w96JC3EJuBcwNGtbggEd4RgX9oVr",
		"12uMvmgGQPzANw1ZS9emTbHwwxScWRprrFoUuh1DarSRWQvNMHJ",
		"12uezmC78wTCgTKqQ6HY5P1q5ANZzhghowAu6tNPDZu6sF2kvJ9",
		"12vFyKeCyAfi8n1P7xBARvECax1zDvwsy4yoGdkQwx5Wz8WfW8y",
		"12vGiSYWiiV85YatrttPMxF1y133ngPK7SqYskjrNCCU1rbY1pK",
		"12vWyihMowujtZuxFMvmQhHh6pf3GgcTPvZ5qp3JdzYhpbFbWMG",
		"12veDZ7XAjjyy9wTS6EToWoM2ZRPqXBcBX6mXjH62bh5QBAHpr",
		"12vmEV5z52zDd5EhQm39zjLk6N1GC6FMSMFptVCuzbQebsrLD7t",
		"12vmwTtd2zztUku7vKkDyPQiHiEGfoMhkpXfRSgk2FzLPTpobNT",
		"12vx41jTsNFvy8SNhaCgwQh6Rag68zKGctLeaW5m4P3LNsDY29p",
		"12wFGfiEpdmfsrd1yavE4QZUrdEaawmuggaN8oYQY5wfqPFu44F",
		"12wKBxNbTkb5dvYxX6odCHHDFVD51B8Bv6k5GhTu4LnasnpXWTU",
		"12wU1kxZ34eNjxHrabZYQZUiw9hX9UkPQn2x1xJfFhYXsWS6BiP",
		"12xb69A9cEYux4XCfeP14Sk3o9rg7SzPjCUssDwECHfuC9Dzhd",
		"13AMsxQeh5mJfxJCnkK2rNoFP8WzqQpHfpMSwo8qYEQFBicpzp",
		"13uv9D1PUe7LpwzeU93n5dHh4YtwXPu3R461WMtVpwkLHmvQFJ",
		"13yWq1vQGw3GnYhXXBSvsLDFdoJNavUhJSLEFvRRGfUsGJisNt",
		"141D9D1xqPoc5s3NUgPhFSdCnvZzmCamEy9k29wgFy8JA4SqF8",
		"14JCAh5EPsfG3vfpxMYnjpKTi5HQ8k5ZZTsSqqfoKasbzz6rTr",
		"14XD9J2DKm7LdMP92wCstFv1YBjVWi5n7cWe6zGt4jUA3Ss7Gv",
		"14c4sgei8mdN9c4RbWQ86PwsdKW69qok8hFq5AMbdSPZFj1oiX",
		"14cNecHmi2JEos4LdgurPpQ6fnSmg4Qx4ExVUQ6f5KkYCZxXBF",
		"14icaw5QUfaKCKk3ncxZePdyhEAGF9M4vgP4CPBhnQJC7aJCYr",
		"15TnWcAuB1fnjuA6BDUyTYkpbvDYvTJR8RcrvtfzYg5gczGJot",
		"15XYrdNQuzRSQqEkDV6iw1qG4b4ZZTep81dfC76H8azpdRmgTL",
		"15e6vjgSTDhPADSCn1vReLJswt4Pgqa8S3cuhtjWWSwhKRyC3a",
		"15oqSpvVy5JPh7iM5HedXektYcyTdV1at4gmgfRKt3mKgyHvv4",
		"15qWQtq1C5BJ2vfPLW1yUb5TRXX2zjV2q7atY3ZJnz8Hy8MZkg",
		"15u9sV6cmv2fkVS3h123tgtycupntoWYSuLCjSmE3USjcdAd8f",
		"15yBBprJsnWWCtdcQ8Gx8KC9sm5RAUaRTKJsHkBVw76c8qr23B",
		"1639czxAeyQtQKYDSrW473VJDpjpfWPgTytU3zViCeMKSyMvEF",
		"16VyQDT1cH373okAWKiHqeLrcyVf2jNL4jK8FzgPqeyM3m3B6J",
		"16sekQzG8SD71f8ZRJo67qs2CdnFEfHvj8yVEbnsYuXmeuyVDg",
		"16yZhhauZKyXyUB6Zq9JZcm21ba8KLTfsjfMBsvXygToUs2w8f",
		"17175sK58sxLb8RotEmDGDfMEx2p5cyV3CmtHyeneiCqSx9w6S",
		"17Q42vzAG8dLZtJsoVDPCTpuoTHa5nWEB9YQygzjENAh2S4QSR",
		"17Xw2rR99QVYnhE8aH8bkC63JP4S62vb9oAEzbEQ6w2DwNCVvD",
		"17fWg1xpyrTwgwLyz2mJ7Ltb99Yd5xPZjvvgt4JGAjL5hi4mKr",
		"17y5PbSuTkxfKHRuMeiDihvnEWZ2WTmDYPRRscj1562c1v4LW4",
		"18DXpzowVtW4vFAtHf8fWQ1yPg3okmjp3LqbwnFBWwu9rGuYXH",
		"18FBL1UvNvYscGahb8mua4NGGWSttLZqwuhmkyYWyP6j9hZJwo",
		"18VcmeAR88sSbLsYcv3wvFKWimozj3B9WKAbyu9ja2vrio1Gnh",
		"18cRtNAUrEcrKQiUf1jmWMHcTEMmdzjrUuw4nnfWdVEKNi8kJ3",
		"19dFU9jDCrgmUPmeFU6ekBY2GFTnmCCpqx7abwooxL5KK4LDAe",
		"19e1966312ucaKfFNcAtHnWfXtN3gQAoYmZBic7KJQR5Jn2FkX",
		"19eCxN5KSCYwbgESukgR7YuofHLHM3bUkPKTgYpSnz3daqBAJL",
		"19pTxRLaCF3gSwURoUpRpJcEot9nVaqs4MhLT11jdqexi2V6kT",
		"19vA9HnVxk3fob7xh2J1fqMEHqhxaEydVZUs6Hrv966uVqHx2p",
		"1A2HTTKiTk1bgArBZyorEAiNVe6bQaLrfW1RPX5cUZNfVKb5d5",
		"1AacLNg6HtrsUVKFzZb5tRGzCYgCL4dyEuPMsR5bUXMNpxsMYw",
		"1AvGif2VuxGoZ6MJL5fpf2oqq5PCgTkfgcihR4TrxjQmqtAXWz",
		"1Ayi4wyaJpcN1yFSUFnA4zxWxS1qfUu7BZ22pvioNVKF4ZR4sa",
		"1BEjR3Kn1vAAku4TxBj4hWZfibq6XVMjXcigaVS87Utq2yPQJW",
		"1BNaQrXdJHW86myhV19wAoKxvcCWgc9S4u8kneYv94D4kSY7tv",
		"1BU43itiK6svTx6thfusSoiW7Mgzn5uCBeh3ZVgs42UBaCWMnw",
		"1Bt4cQbWvEEiWZWsPYJijYEaKix9xGyH6Zd8QMKDMUuzCBVvwP",
		"1By478J27CKLhu4h8yEgsSW5moogLKDQHHdeT3gq3j7ANkuuyq",
		"1C2ofHQ2v2YV7XvvQTGTFTSfvn4LYLJny6ujLVUSZCQ5928VWm",
		"1C6v9hkLaMsHpB1jHFdm5HcHgDDprFWkd8v5F5fvqUJbwZ3NY6",
		"1CbmNbfGDqrSFHEkYcscs3JLHBmdzmsH2egJndQoWKBigx5URY",
		"1CgqHFk3QwM3NZ3x7g7ieb3sL7XJQXfAfdBfBAjcCpPBmJojyi",
		"1Cw25jN2UYGtGnkcEtwKKP8wfBaVSUHRL87YiCu6otjV1CPJJh",
		"1D2gJ2qHsDcpt8APLJQyM4vKzvtgwPtpTZ1aRNma1Qtkjws84D",
		"1DB8oZvowGqrazHVVWsxGf1v2D8rAUzLHV845r9vPP3p5nJVpo",
		"1DFYYeTQjWEMbtKTnX7YXq6ac5jxGBAVppF9bbMhf3pqmMoAhC",
		"1DGs1ApAouffht7ES9X3YtSC251JSD8KQCKfBg9tvfcsM6Bs8t",
		"1DSxhHWv317a6iDQd9wPbTmbcuvJ6HcsgVuPhpUGKKQdwC5qWJ",
		"1DrPsQBbofTkWcEgdpKUgbmkdZQ59nD4ZpFBfeLiW4mUdG2skz",
		"1DtgzpAufhwpBvHWTaoG5a7g4P5Ma7Wj6htzW3eAUbqbsrWqqp",
		"1E4dCWpL5jdgaNSSM1eP3A3LvfFe9zsE4uV1KynEqATd5589wV",
		"1EL7rNJspwH8gCDSZeCzspiDN3K88RLf8rsVwuzYrrtcZqMLQB",
		"1ELE9X6HbaSQxPmJGgWTVdPE22ThD46M339k54kwAV9kkjwuUx",
		"1EZytF6HeD66b4yyrCHHhv9j2pahiAZPaUPrrnNozDcoGGXEDU",
		"1Eik6aiZebdNXeBin8nz2yst46xPK7TsDd9Kx1VU5NtHZnz9MV",
		"1ErG83mjPc7n4vWUioxBXaWq8EtGfDn4b6UC5avdDaJeGGKRkk",
		"1Ez5YZ3CMaTQZMUQR5GEizRr3sjHM4wUVQWrwDTodAVb1vcbYx",
		"1F4oQqfaB83HwC4Zy8poEKPG6h7h1gdxN7fPE2gy9LwUzUvBeB",
		"1FD9Z8oXEpAdhnZBFkVoCVXE18YNL73L3215T5rAb8pmDLUTQz",
		"1FKBxXjbQwKCc25ViMWeRT37jGKCXFZWpyCfdPf297TFfc4bDS",
		"1Fzo4FNMR7rynrSc2nhjN61cNdkFybLaEjTdZ4TLocMLutEbVz",
		"1FzpTp5vJD7yRumFQnsT96cZd42ifAuPok7HVueLK87PnyfbW6",
		"1G31H4TmdSv6sDubnLoK9WGUJEnDK85dsf4WRR8MqDkGbEa9kQ",
		"1GAFrTFSEDDDG6tkSMTa1fXb1C9aj5UAGvBVBofyGz9sN53xn9",
		"1GRhVZKJDKSMnPheq3ufQn1UxDbAsbHHBPJq5QR94JMKVe4yd6",
		"1GkFSKoZFLdNu9PktskxMZ33Ygcdo5iF92HqWHWKTjFtriyDV1",
		"1GtVkRoq35TQXSeVbTM7awtpij41TERjBQVKjLCYmG1QVL398b",
		"1HBMK5ZW2WwZqJdfbcSB6xYom2Yst8qM1ximPWBsK3MqV5ZZAL",
		"1HEHiury7NMbZmxYknC7vqhrcfDcbyg8RN24Jsx2V6oa1uRgwX",
		"1HnSuZ5FP3oSSpyKca9QZsVjoxKzSGQYq95DV6P9Ctu53YDkQ7",
		"1JLrCZrAr8AJGKKZjxMyeKYAH7dsK38kf7wuCThRoE7654EWAa",
		"1JN9KRgiEzFNA12ZEEEqU4E3sMJfZwHT5cenZ9SP5sEg1ohjJw",
		"1JStbrFsmtD92qr8QZXhJGSrRpb3us8XzJEjb9L5Fd7kSpBAtP",
		"1JTsdv2TRAbKRSPCt9YciCA8dhtmyzdfSMV6bDXMTLiNX4Kbhh",
		"1JdP4MKv7ocnW8DypjEvofGc5nwG7tUGvEeWgHhPqiimYuTn6N",
		"1JdwyuCfd69BEYsGAw1nQ7gEfxkDenZv8yx9QqW99oTnvCyGYF",
		"1K3QrgVbtGPDgedm4Dd9VVdouiZdDT7d81W219o4PobS12ZiTn",
		"1KBjbhjXirjPrFWPFcx46idCjSfPj5fv6r9ZbSn1eX6YR6Ju72",
		"1KHKFDpogqg47dwBDjniRj6ognFitKf5QBqEP6chLBXk9ZEB9q",
		"1KcSqM43e5wvhkPgMMBLtygiQzxw1zimf9CD17svhjpM915bXv",
		"1KsjG1jQbqhcLyhYDYVVpTBGjRDT3FaapXs2QFs8vFZdv5XrmP",
		"1LRP1CT9sMphhqoXcEmBXNmyM2mwWx8tso8KJYRFJXw19XDbq4",
		"1LUADbLBh9jq2Q9sDnTp1Ysk7WW9wfbpSgA3iUDqtzoxPDhKNA",
		"1McfbhsRnYCuUegjEDBSAfZyBAbZoJ7FHLsPqqxFG5UvHAPjGb",
		"1MgQTn28QgN7DPwu26NN43NAroKL3WU2gzHsXLnYzZVb8ga9NG",
		"1MpA6WycX6YVQVsUcAbuWzYw3HRJbjxQ2mzmqRYzxPHZLh6rdZ",
		"1Mrsf6uw58Hwsg6PZS27wx6zNZqBV1BEdnw9Wkvb8qXNHt214e",
		"1NDiyYYg4XGxwtmeKVXMUULNzrRH7o8Q4CZ1nR3Vs2P7GSVm6R",
		"1NaTFzyQMoeBFSE2B5D5AjfwVJHFYyv7Zo3CiPeJocKYAFjXxQ",
		"1NbbBFKAPpkwGcPHrAhQftPizYGzjd6hQHo9ehREqdK3DKk9LE",
		"1NvyJQUQfoNyU2M9DB3T2s2mQkyhvfHZ4oY8Ebqkd95d4Uion7",
		"1P9KHcstJdaurnv3Y7fUsH1FAsWVDVSbGAH4dubfGoKPQDkJUs",
		"1PEW5wY4dDdmGhG4xKUh3uUe5cSH18riXV45CboiYzoi2NNk41",
		"1QLHG2Abw4ZXJK5dQKC5A3KZSj2C8ZMXP7cse5Mxqzgf8VxKPh",
		"1QjZ4E7xgft6zd1TZTaFk1Ux64Q9vtNNunFGcrNP6fV1SrZjJK",
		"1QtXuDDBNc99s93sZXvnn6rLkLJ6fa19U7L8nSLUiw8xgch2Pz",
		"1R3Enmzf92SsezL1cMRZx4wswGNvZtwuaTY2cxMZfX21rRzq8A",
		"1R7ME9Mpr51oCTY9dMAjYNNUeNJTLsN49C84ygJsxJVznQcUPG",
		"1R8j8uy1bq4XoLcbnG4EeDE6a2eVDRZ5jghJ3pK6KH2QY3zQU3",
		"1RFZtKDxm852QsDA9vaHoaZTLT2S8eEUw8AQPqN4gEPakLMnBU",
		"1ReobmQXfXbjn2hs2p5Ho7YwGUnX4JqEF9THLLGTK5axS6G8Xy",
		"1Ro714tZc1wEzttYn6z2jt58guTDSF2PPWazxQX5kvJH8daCGh",
		"1S6xJQ2dyo66NSdQJ9Sr5JyG4uZwWMksCJsM9CENTifzXPtbF3",
		"1SWuPwFoEtk1F8hnwnE34kb4FRrjSKb1rLuLJd1eC2E39TsYcE",
		"1SbhJRzPi8yeTzsgCCbq5SfiaFizNoFB4EektCVLUih7rrH5Vy",
		"1T2SRd5w86Bz9vSiRYBt9qMRHUr23M6mAoDd9mgpJDkBe96Kzj",
		"1TDf8BgfnxXPCetDrPdqpA9bCrVBUi4FWjwN8FbgYhmzSqjLWw",
		"1TLSd2J1vAfMTzzCyLM1LBDF6eMGPg2tpx3wXfxBjJSAFzc8ks",
		"1U4KJ5DekDgXybv6v7WFYNzWVRNWCMt43njbCAvXpputC3VW6T",
		"1UCeh6VHcQUmLfiKj7Fw9FLsoDeP3vyvUpP5JttfuUg2KAqPfR",
		"1ULT4bgzQhESqhre7GjJtkpn51P2DXBrfFRiorHLahY72pL2tH",
		"1UpVWSaMJ2U38kJcnFf7V1TAjpgu6CvDCAuRCrdxZHunVMqTeJ",
		"1UtGgcJSYRekkrZrT8dF3vgu7xYUaY275GTH6VNXbCgH6Loi8b",
		"1V1q6GakpKXP7AHbDhMCSWpUwxAyoAkpnUoALqTEmQpJ8WniED",
		"1V5oHWrT7F42nYUKpMCq211YVGBNJAvuQRJK46BAbH9uvNbU5u",
		"1VCxCJ47gRRswYJMuVxxa1prS5i9SZFMJPtfLrgrwQg1Kh4yQs",
		"1VDJJEnAvn1Lb75vsy1mv1UFBX5wUSdx5az2hJ7DFKgTeZ6iBT",
		"1VLZqXffULrvffvQJniEQLfVM89WcFnvKKLsa7eqE9WnmJ821S",
		"1VLkjmtHTECmtj9QTXGBK1rZcWbAxhP5wr7L3hwcyUVcAAfhQQ",
		"1VNja4EE88DsMzK4drjtXkAebJRrzi29yE4gt5rPhXuJwuBywh",
		"1VTJQDLYRzY6Fny5JuLLHNPMckYYFZ1s5cCprbcoE9VqgEiW8F",
		"1VVQRGZFhMyp9LRS1yjyXsW6ugMqtHaYyUYMLutRhy6ASF7AWf",
		"1Vn6aJo1MMmo3fZHC1tLtttHhpZC9TRFjorqtyHL88VMA3M7Fq",
		"1W4A4qGcSQVB8XKK5KVGLSWHzfDtJAApMUYkyqAqQMtLFQtPWn",
		"1WZDaxLyeRvieHwWpn2qKq8ruZT3tNzyjrWYGVPgMiWTLqPQa5",
		"1WdNm63iKBi4W52PQfqHnmhZSsqxh7LLiMZuj2kshndJh439a6",
		"1Wo3x7L3Dge3CiRg7qm2PDPkDPLJCnT9M7Nuvv6qkcp5hgXW6J",
		"1X9u2NpdqYXvygZnvczbgi9uD22xAZMBJ4gQaGqNWB719gb9H9",
		"1XXB2PNKVZJZQr474yt89nANBML1JwRkoHhkxdXvnnVxKA4Zee",
		"1XhnVYJMJ2ZCPBa1EQKPU9zEFGw18TTkFRgzcuVRGuE2NJ8xFe",
		"1Xt51o46FNpW8ThiW1q1HECfx1Rtgwvo4gx3rbuCMFtycNQw6Y",
		"1YFFF5pVUXeUmKTJoeExpgVxEqJtWZPpEYZK4nfuWMKAXKP3WL",
		"1YLK5fPRqrRst4kvrwMKRbYdDDEAk1oHzXhxYB1BTKeD4Qb6gm",
		"1YLoKf2CxEKJAFR7QLdZ1c588KoGmvtsELW5sq8Kgekaxy1mY7",
		"1YQbxGis1Fp2VQK5632rVYNj7LkFe5HJd2XuYEjER8FrmA8Xmt",
		"1YRBdMMwBCNDJvktGbR7uL8fv5RkUuBJ9daseYid81JzvQBbj5",
		"1YgzNoi5HbJ2UtFbsUhyDHNjdeo5yMKyQvqMX9tcp26Li8mH33",
		"1YquAnaeAVxwdJbc9zgca6ccx62gxF5YXtP4yHSURdVFZPvin5",
		"1YrNWwnaKySi1sqJCbjCi3d124cCCCobyn3m4PXrhnjevqFRqn",
		"1Yw1dgJrWX5FR1DMcgtHk1YZAfC53obHQmb7JXXd4Enw5YScDD",
		"1Z696kvzYDUGRgxpx2P57WASUWp211hHEcsSqGQJRC8TXcsAJd",
		"1ZJh7xSuneu2CkmjH7J17jAcEKUYW5s8PB9CtNToRukSZZqeeL",
		"1Zt2kobzry5DxuMEVvVPLAwJC8JNgjDphHDCkSGYqSFrpikAtJ",
		"1a6pUEtTXERSkwQP3zM2A5N4X86FpqEEsha5X43KLrSHzqeKHL",
		"1ae5GW1jiq3naxBWRc1wb7CCtZ7Cs8CtFFxtujhTSP4Y6hqA8f",
		"1ar827w3LihxaRAcjo6chLsA7nFETaXkoQUbXqsX8yqEsBVzcx",
		"1arkEAxZfKjsEUYY6mSXdRanMGX32hZG3Ao5fcWbWohgTKfTxz",
		"1bAKJK3nYaGLqyjGtUjFrVde4CkfrMVNwS49JULbTxdbmgN8zA",
		"1bddjTGbpUPRirk9UzrsjvsFSottXc7G7pM4icV6f2ULofv7mX",
		"1bekpzaYuDDDCEGpZkYpJLtZNcfE8EK5a1wsapPKxM3TsEsgYa",
		"1brPD4Vf9NHgCNxuYGoFKuVPDU5N6c1n3ih5ncU43ZDcH2Xw5y",
		"1buYyLy4X1dgnJVkFzes4CBW7wfZXEHuqe3CzvXqpibRHmspwS",
		"1cRF7ybJ2Nro7LeCs71GP7VcR3j2RaCiFE26h2TPWRkMizr5iT",
		"1cmUzpNW7vrxzhft41yRfvzpaZedm7W9Zir4zwCkDBHx468HXD",
		"1crQavsh4KmZ1CZvj2RYQfJ5wKu76gArS5AmiwgRDZZtP9bFyL",
		"1d2ySoHT4jACLzEFJCWWGnxDdmWCHBbgwdVzfmAvGyUneMuZAp",
		"1d9sdjKHHFyM8DGwYHpFLcUJKsNgt8nj5iaobTLtfpJRqve9QR",
		"1dGmc6b6SfUzMyVaCkF9An62wcpfbg6cfV36SNVqYZYurEyRqV",
		"1drYYiXxYySBHE8hexfzLSaGpYdHgKZBWaRgSxzre6kMzY123y",
		"1dwiTwN5Eqx5jCk63nG4gkqJqDXAiET2GrwnfxZGGu6rTY7Jqm",
		"1eZBvdXQB8BYF36DXQnsov1SoJ2NJuerC8asY6dhR68fxoLZxf",
		"1ebHSiuaAZbTPtfxS49eMB11qVyhDRh5mYiGadZgMiDxgaVB2K",
		"1ebwuDMTyGHk1QzMZH8pBUE41KhW1eDNGcFZnCobvNAHJuxbY4",
		"1ehuWTYHCLpFKXTH4LDQctaAvQmkh9n32bichACy22N2sku5FN",
		"1epU2cpWkyWbTTgcXTEc3RpFDERybiCYXxeSQdzVstuYQi9qYv",
		"1erDK7nTN32bqvrTLxz6TkqHTtRvFy9JEvosWxiAhfuu2miCaK",
		"1fBw3X8LAMBaReJEUQa5TN34uRS5wZgU8iCniknkFvaT3BUXJs",
		"1fEUcV3prCWCDNN3n4XiXuV1CzKztWsTFcLjA71djYTRAtLzam",
		"1gFWf6BtFjLdRTNo2yZ5kdvmBeUh88XrT3H2pVGYzLwwuhxNkp",
		"1gKUR2u6EEgxrXBhQugmWaRpam63F6JXfJduS1K98J6wGc8R9a",
		"1gf6kvvi8r1iMYASJoFyJXGNr8U8aHWSkicMNLJqgeLJH7BujM",
		"1gyjJfDkJ4ZmtDcthpZvBwZNMaTSXkwkaiMZT5bazqSHNsugWz",
		"1h7u9jVkCXrEutUvy3LKR39VR39sbp12Q3uSAYfVB4wWCdB1Wu",
		"1hNMuT5cSG1CcB4Nm3jvyFAwUpn2ZDy6Phvf5FV5iqS6xmG4B4",
		"1iVu1D9y2uvQn88cfQ2qVmB9rLC6J1X3iGuwRF9b8dbgJ8DDEW",
		"1iYTpEBbDtcmfwh6M3wXPtD1B8o2fGdJcxy1r578FgntDsRpH9",
		"1idGzuC5E2a19cmcYy9ik6rByvf5A288MGei2ZjEQhqoaTJ3r7",
		"1ieKKcFVcQAWtCvfUbfs6Ee7YgTQb7rwExKs8WVnRRQY4E9fAb",
		"1j7J3htqupv57JhPMaNdAXgmPQkZLYjZkB4cfviLYFjH2kiJ31",
		"1jCgCMQGs9DpgaxvZrPRUwbJzWBuTzftY4tbzQhtGFjig9hGyr",
		"1jWDkjMjXjep69PBcPi3eV3BassiGusQSzxLv3NwZxryumY7Rr",
		"1jXnF4PEqE3yxBZpq8W3VKEgDoM1qiRfpd4dTwrd64ZMgZCEsp",
		"1jg9YT2viqy3m2bPy2mkuv5hW3CapjpeeC58ER9ZCsGYDNkQmm",
		"1jizkDHqiUeMpX1Dt7BdXxAyiCFNWGJyaXFfYzrSqg9kpNp6Pu",
		"1kA1eY3Sgh7DukKcq9QowBiN4GbqDicCzMdw4U6Z1yQ83UX7Kf",
		"1kAjJ7mmiN43LoncuxP2dY5xDCFSVMNeEeNpXp61ZbXdp1QA3z",
		"1kK2j8DHPxwGs2A4QRsDinE1dk9X9Lx1hsw3Jf87xyU7f4hjMK",
		"1kNc5UPHa3StZ3WeKVg5362W6gwGkmch19B4q1EQtYjTHFDveD",
		"1kXAZoxQp7CiDozLxktXtQsw4HsxNxBCydUiBBc6Gp3n6VokYH",
		"1kijEojJvmXYCnfA3WpUZHayVrwgKctjQzf6qi2y2DWDi7UTJn",
		"1ktZxvzrfYpeuXabzCw98sQzeH3wFP9B4ncq37dtA3EJqN418L",
		"1kwwu4c7jezdcmCRNkwDevcZgjwzYvgQ8vUtKUmQXYFRLvX9e4",
		"1m5mcdZWqPeUPzbqp6hPDgC8xPNHHXzJjE8qzJjJdKbEYt1Day",
		"1m9G7tqMP7Dq1XU8GL5Uy9CCWc6cKRDBtiLefddahszNjDQnEd",
		"1mEsuq3wz23gMRcN5xiSVW3No4kbvoAh6LDcE3L5vt487guu5v",
		"1mNWa8Fs5ZbNsBLhRRgtCa4DLUs5kLkSXYRjhfBoDcyxqQfmnw",
		"1mQzAP2LxAxhhZepLNa9GCz9qUdTADh47GijK3nK5FKzdcgs4V",
		"1mRKHMiz6S4Nq7naJe9Wr8JAMd28hkQq4CbbaA5MZN9DzESSec",
		"1mST43Vh61PGBvFCf95KAijQrrJfi6mUyXC7ZhEa6dGzKGqcis",
		"1mb96PkoLvYh8LHxmAZTUYHfwFwMrW42L5W9EpWT8mhaJVNutm",
		"1md1tdscbkEWNbNdDtUmC5DTYz6iCKuTkCnkVd1uuB4qVTxptY",
		"1nQ3m714j9kMBpAVVX8ujntPvqMKkNqb1cdrgN6fBTModP4DL4",
		"1nQfR5nFXK8RYkpFV6SzeeTACNSHwNkQpmxyCwAEDXYhayyrUg",
		"1nhidCSBU8yZ9CDyeJ5B9uodr6y8toXSed21H6oiLzQjfhFaRN",
		"1njgFv6iXMiDJr4oFNjuusMS1XTQUYYbkHa71CjZtEYnwvpvH",
		"1nmt7PkkQMKmqVHhFeUeJxye1WvjgT1waXRZha4Ni3VNPbnHHo",
		"1nnsZwQhpsdMoypwBHjBRMDgNiymYzkkrSn3f4E7j1e3ig8ngP",
		"1o3Xzscwg5211n47iGtzBmEppY7cinK1Aa2rUUESGEkxXba7db",
		"1oCiscMVUBSmJKSLqKEyzNpcxCNv23LSzAMHnVt9EdxwzrRJhp",
		"1oJhCNJkLJ9xXYGfjowCnsnvTdakk4tw6vagZoZpFqMZS8SbbV",
		"1oNhTNBtr3sLm3i6krWDkYUkRmrXqEtZMCqmK5Pv8uJomuJz7F",
		"1oa8U1CFKZfPVzR26i8uPEWiWhafnDmL1pdzmkU4jGerTEw2E",
		"1ovU4dDiNqdR2GfXPxcuCW29hC6qRWsM5bRiGzJk8j7NiFewjL",
		"1oy3GNNdeW5fzJWe79omA8Ke1qvVVReeS3G6nKLQkiMReu8eVZ",
		"1ozUszutuu7xXD1Muvn62fDsPUg2MwwCUzc4BYoxyHhikiRS3m",
		"1ozbr9V5haBZvhdiukiyGfMxUdnwiHK4smQVG1My2K8aGhvtu9",
		"1p4h2jK7zNs9RPoi6nMoPQWoewAJzCosDU3zrYaiZLvLtxAvDt",
		"1pF1W7P1Jeyt3PNpMQ1ZS9nvdtyqKU63pJcwEruZZyGnowhfcC",
		"1pLwZpPaUe3VeDrnkz8morw6SWQFZuRTrdh71BXJczHsyhVYkH",
		"1pa9DCg2Ng5ydC6RnVAhydSGrA2DdXW1o9dHeh5AReRbgLnPQt",
		"1paSDyW7dEpKnEpoJyUrsTY5cPP3i4mS4nhoL2685hjmVmnoGc",
		"1pdUoToGgzs5SctqjH3zvV3r3DztWZjTZHz7BkNwktLwV44PE3",
		"1q7XZZ8mP2RymyEJZvRogTNsjyMcYdnWUB7V6dHKWAAEvySdqz",
		"1q97YUj4dhpt9x1CXTGnSc4uYmSztLTKPCL5PCy86VcNhjnVF",
		"1q9ip8q7NFih7E5ZVrZA2cUJSZ5xeLYSvVUzztqeouj4cHa9qB",
		"1qWpj1NRBn8mrSWrLHUqK5VdNgx49ygMXNFRCBQTdYfPiCzV4W",
		"1qgtVd7BwgfaeouxopPEMiYKeneYGr67zd22teEq97uMxYVsVs",
		"1qpm6dT4WMq8hqdxUyfHScbbEyACLNeVyYGCHKUZnQr48RKsgG",
		"1rPckYmxjPT19B3nxZR3SGAvkNqcWUXUbniqPHQi8bYCs8wNjU",
		"1rTsXUkuSVkenGwX6TdX4zvz2UcHiR648wgC5fNHct8bE5GuxG",
		"1rsLyfzMkaUHxP5zr972LAEf6iYLM6isusJ3dHbev6bbZ1e7R7",
		"1tdUrk467sd6XoXKVmks8YrN41NUPHYyYDNtmHw77vSqycJRGJ",
		"1tpz5tieQ9S1zim3fzwUTEPesgo1RvbnoamGcbUWJeA9z8HEjv",
		"1trtyu4v5FWBRBaqV7kMuorgpkiBfzLUQkiemjZBNxXCeuKkx2",
		"1u2ny9PVA7JbAYkPgUwy4Xt55xD5SXnw922TTGngLfAFdDmT2e",
		"1uEWx6VvCKPsEsGtwZDcMenvtzxrsVGR79U7yx44tA82paTtrs",
		"1uGYFhZAb1idsv637pDuPtUCygPzAzHChcyv6UY2N2g2tAea3r",
		"1uHdhkAMaAUyq6iz2sw6MYTLdYQGcc6a7T7yB7n5WawNzGk1WF",
		"1uMbRh7eUncRt4CAAsQuKpoKmepe7UztTwi6S4zFV5CpnKbGUA",
		"1ug9iqpRDW6GUkL4bmoU9QbEZyee9iXCEEcnmVECJjhAHTFX8D",
		"1uid8eef1NeQGKLMJa6aUEYqJsNcwVrQbYBvst2y7Scod2pdaB",
		"1uondutMYpR5xMqkqe3xDB2Z7svjmC7hj8yrxVJPoup5RT7aip",
		"1vDL433A6V19gpSZAo1ew6dCQekSFacGNdR8se523WdmQSHahj",
		"1vSvgRdVSA3JnHPy3MmnrzqvjaFB981wxMt8kHn19KvznozUhg",
		"1viELYXEF6dgqEk832RHJfRod5ZYHsdpbu37fJRjwJ8xokqsDu",
		"1wUqfC5eddaf5JB9eLG2u7Zk1ANiUwy1TMCiDRjQ3uxHAJ7gTE",
		"1wndHPDZf9p18B2j4QpZDGPwpMkTZgZTTrnEK7esYu7pYnKHPp",
		"1wxnjWeLgT59dERfsyfuuNrU8tLWHZuhQHoaSg93DXCUdX2yxA",
		"1xEb3RW4TJG3C25AgJHHAktmzZauFpCBFQkXUNEGZ7AUovc7q2",
		"1xEjJTS79ewX7rVRY449QTE5T2B9buyc11bVr1N7GEzPBYyUd8",
		"1xQwSVGXkfdDwnyEMsZv2b26RxGgPrN2AfNCqobQctMHAimKZ5",
		"1xVYcbQZMHVxq1HqAGhczUhiKmox3oMBsZu2XHj2pUdZiGjVn",
		"1xVgqGE1191UkYhC8pFJdKXacYnDXaCGskjf4uMPGjJypTFYdZ",
		"1xfYYx9WmCzoviKqgae6M1nYbS94xKTuTGmKrMEDFTz2zQo9wW",
		"1y4haVY1q1d2saTjygbSfSrrrGHLE3ma21T71jcKRRrPJ5wQ7v",
		"1yGA52wd7RTgP8UbXWnqYPg33cLPdNN4Rp9EgnjPLRmArUZG1C",
		"1ywFQqCZNpcPDahY9y8vsHWDK9PuocvtmBmFfuKhGh81MK8HU3",
		"1ywRMuTaVWYvquNYurGzE9BkPwxiuUyZQXqgPBTwezhABJxAzJ",
		"1yxBFUMZDay7km3bqcRYzD5EMWi2WmDDiPAQHYZobvB11za5vw",
		"1yzcSLrpmDe7nY6Qa32edjj8RXgwT6hoXVwu64gP6RqQChzfSP",
		"1za9uf1atg7GaGHU2syWKFNm7yZMGwYPNsCMRaCFAzb4TssmZn",
	}

	warperDBrrTest statedb.DatabaseAccessWarper
	initM          = make(map[string]string)
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_rewardreceiver")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBrrTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))

	for index, value := range incognitoPublicKey {
		initM[value] = receiverPaymentAddress[index]
	}
	return
}()

func storeRewardReceiver(initRoot common.Hash) (common.Hash, map[common.Hash]*statedb.RewardReceiverState) {
	mState := make(map[common.Hash]*statedb.RewardReceiverState)
	for index, value := range incognitoPublicKey {
		key, _ := statedb.GenerateRewardReceiverObjectKey(value)
		rewardReceiverState := statedb.NewRewardReceiverStateWithValue(value, receiverPaymentAddress[index])
		mState[key] = rewardReceiverState
	}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDBrrTest)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(statedb.RewardReceiverObjectType, key, value)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	return rootHash, mState
}

func TestStateDB_GetAllRewardReceiverState(t *testing.T) {
	rootHash, _ := storeRewardReceiver(emptyRoot)
	wantM := initM
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrrTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotM := tempStateDB.GetAllRewardReceiverState()
	for k, v1 := range gotM {
		if v2, ok := wantM[k]; !ok {
			t.Fatalf("want %+v but get nothing", k)
		} else {
			if strings.Compare(v2, v1) != 0 {
				t.Fatalf("want %+v but got %+v", v2, v1)
			}
		}
	}
}

func TestStateDB_GetAllRewardReceiverStateMultipleRootHash(t *testing.T) {
	offset := 9
	maxHeight := int(len(initM) / offset)
	keys := []string{}
	for k, _ := range initM {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return strings.Compare(keys[i], keys[j]) > 0
	})
	rootHashes := []common.Hash{emptyRoot}
	wantMs := []map[string]string{}
	for i := 0; i < maxHeight; i++ {
		sDB, err := statedb.NewWithPrefixTrie(rootHashes[i], warperDBrrTest)
		if err != nil || sDB == nil {
			t.Fatal(err)
		}
		tempKeys := keys[i*9 : (i+1)*9]
		tempM := make(map[string]string)
		prevWantM := make(map[string]string)
		if i != 0 {
			prevWantM = wantMs[i-1]
		}
		for k, v := range prevWantM {
			tempM[k] = v
		}
		for _, publicKey := range tempKeys {
			paymentAddress := initM[publicKey]
			key, _ := statedb.GenerateRewardReceiverObjectKey(paymentAddress)
			rewardReceiverState := statedb.NewRewardReceiverStateWithValue(publicKey, paymentAddress)
			err := sDB.SetStateObject(statedb.RewardReceiverObjectType, key, rewardReceiverState)
			if err != nil {
				t.Fatal(err)
			}
			tempM[publicKey] = paymentAddress
		}
		rootHash, err := sDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
		err = sDB.Database().TrieDB().Commit(rootHash, false)
		if err != nil {
			t.Fatal(err)
		}
		wantMs = append(wantMs, tempM)
		rootHashes = append(rootHashes, rootHash)
	}
	for index, rootHash := range rootHashes[1:] {
		wantM := wantMs[index]
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrrTest)
		if err != nil || tempStateDB == nil {
			t.Fatal(err)
		}
		gotM := tempStateDB.GetAllRewardReceiverState()
		for k, v1 := range gotM {
			if v2, ok := wantM[k]; !ok {
				t.Fatalf("want %+v but get nothing", k)
			} else {
				if strings.Compare(v2, v1) != 0 {
					t.Fatalf("want %+v but got %+v", v2, v1)
				}
			}
		}
	}
}
func TestStateDB_StoreAndGetRewardReceiver(t *testing.T) {
	var err error = nil
	key, _ := statedb.GenerateRewardReceiverObjectKey(incognitoPublicKey[0])
	key2, _ := statedb.GenerateRewardReceiverObjectKey(incognitoPublicKey[1])
	rewardReceiverState := statedb.NewRewardReceiverStateWithValue(incognitoPublicKey[0], receiverPaymentAddress[0])
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBrrTest)
	if err != nil {
		panic(err)
	}
	err = sDB.SetStateObject(statedb.RewardReceiverObjectType, key, rewardReceiverState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.RewardReceiverObjectType, key, "committeeState")
	if err == nil {
		if err.(*statedb.StatedbError).Code != statedb.ErrCodeMessage[statedb.InvalidRewardReceiverStateTypeError].Code {
			t.Fatal("expect wrong value type")
		}
	}
	err = sDB.SetStateObject(statedb.RewardReceiverObjectType, key, []byte("committee state"))
	if err == nil {
		if err.(*statedb.StatedbError).Code != statedb.ErrCodeMessage[statedb.InvalidRewardReceiverStateTypeError].Code {
			t.Fatal("expect wrong value type")
		}
	}
	err = sDB.SetStateObject(statedb.RewardReceiverObjectType, key2, []byte("committee state"))
	if err == nil {
		if err.(*statedb.StatedbError).Code != statedb.ErrCodeMessage[statedb.InvalidRewardReceiverStateTypeError].Code {
			t.Fatal("expect wrong value type")
		}
	}
	stateObjects := sDB.GetStateObjectMapForTestOnly()
	if _, ok := stateObjects[key2]; ok {
		t.Fatalf("want nothing but got %+v", key2)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrrTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	got, has, err := tempStateDB.GetRewardReceiverState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(got, rewardReceiverState) {
		t.Fatalf("want value %+v but got %+v", rewardReceiverState, got)
	}

	got2, has2, err := tempStateDB.GetCommitteeState(key2)
	if err != nil {
		t.Fatal(err)
	}
	if has2 {
		t.Fatal(has2)
	}
	if !reflect.DeepEqual(got2, statedb.NewCommitteeState()) {
		t.Fatalf("want value %+v but got %+v", statedb.NewCommitteeState(), got2)
	}
}

func BenchmarkStateDB_GetRewardReceiverState1In558(b *testing.B) {
	var err error = nil
	key, _ := statedb.GenerateRewardReceiverObjectKey(incognitoPublicKey[0])
	rewardReceiverState := statedb.NewRewardReceiverStateWithValue(incognitoPublicKey[0], receiverPaymentAddress[0])
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBrrTest)
	if err != nil {
		panic(err)
	}
	err = sDB.SetStateObject(statedb.RewardReceiverObjectType, key, rewardReceiverState)
	if err != nil {
		panic(err)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrrTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetRewardReceiverState(key)
	}
}

func BenchmarkStateDB_GetRewardReceiverState1In1(b *testing.B) {
	var err error = nil
	key, _ := statedb.GenerateRewardReceiverObjectKey(incognitoPublicKey[0])
	rewardReceiverState := statedb.NewRewardReceiverStateWithValue(incognitoPublicKey[0], receiverPaymentAddress[0])
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBrrTest)
	if err != nil {
		panic(err)
	}
	err = sDB.SetStateObject(statedb.RewardReceiverObjectType, key, rewardReceiverState)
	if err != nil {
		panic(err)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrrTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetRewardReceiverState(key)
	}
}

func BenchmarkStateDB_StoreRewardReceiverState558(b *testing.B) {
	var err error = nil
	mState := make(map[common.Hash]*statedb.RewardReceiverState)
	for index, value := range incognitoPublicKey {
		key, _ := statedb.GenerateRewardReceiverObjectKey(value)
		rewardReceiverState := statedb.NewRewardReceiverStateWithValue(value, receiverPaymentAddress[index])
		mState[key] = rewardReceiverState
	}
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBrrTest)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(statedb.RewardReceiverObjectType, key, value)
	}
	for n := 0; n < b.N; n++ {
		rootHash, _ := sDB.Commit(true)
		sDB.Database().TrieDB().Commit(rootHash, false)

	}
}

func BenchmarkStateDB_StoreRewardReceiverState1(b *testing.B) {
	var err error = nil
	key, _ := statedb.GenerateRewardReceiverObjectKey(incognitoPublicKey[0])
	rewardReceiverState := statedb.NewRewardReceiverStateWithValue(incognitoPublicKey[0], receiverPaymentAddress[0])
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBrrTest)
	if err != nil {
		panic(err)
	}
	sDB.SetStateObject(statedb.RewardReceiverObjectType, key, rewardReceiverState)
	for n := 0; n < b.N; n++ {
		rootHash, _ := sDB.Commit(true)
		sDB.Database().TrieDB().Commit(rootHash, false)

	}
}
