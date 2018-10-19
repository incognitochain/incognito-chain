#!/usr/bin/env bash
if [ -z "$1" ]
then
    echo "Please enter node number to start"
    exit 0
fi

KEY1="1cXZ1szvjp1Guhme8d7HQwcMfpNfS39S7CVFqoETiFHJ9JN5nbKqCaGkM6rTps4r27H9D3pEmwvkBcGvhvyJusbgbRumJ6KcT7oLob3NGhboFnmsuiCo9z1eVWb2TFxi1mbhruCErCfCsGnxxxpo2YfGndQfvcTphmz1QykqkRucWss9TUnp4VQac8SnmkDfkSsANzLfGmHRxbWkZRQXXXgBHrBPiBBcWmCGANKWYFaGK47ptBq5FQdgoVFkqUd1ydcCkAnaN7iSa1fW2MNQbzniVCCvkycWz9S56bvgZpUJAKSFCiBrspw3gSdeWy7ox1aH4azBFc6xUikPqZ2MgfQHjnn8TUjuD4U8c2HGrQy99F6XPupK7S6KNMoBxzMxoVXSXojrikKQp4nxwi9szw7eJdN7PbmNUUNHkH7zdUdB9vHHGKSS891aHiCjT8vm2DoGnfeiDnQ5Fpd6ifo83kpGDduzoFoGsbAd2R322et1Woi2RoT1zsjuYQdVKSo9G5N616UY2fAT48SBGFu3PAJz8sDoYPQEeBghk58YiFcjBcyUshNaVwffE3QijSLQAtdB9TrAF1EtjypyU1ezzvM8Zhiq6UD1UCP5h24KC5cDkUpW9sCkJDpGJF9KRYK1ymktPpd3T8N9oS6uKUM3VijS2MkYcMkcBVhoPsrZa2dWyHZ1E8SjXnkHBv3twpJvehkvM69pe6jw68ezsJ4TLrUBrNj4Z12dRnEZSuQfHUqdULN96wF553edyt5tgzJ2jGmj8LFbAE"
KEY2="1cXZ1szvjp1Guhme8d7HQwcBVUyF7SYdSbmFkCVpPSz1BfgkaaBykNccsorqohKbfzH84P9kn8qfWY9YnXUFE5p7aMEiqmcuFfp5G34uhAd3VPB53V5P8A8yeEZ7TCbk4QFNngcR9tSxFfQuvXNWycoSdZcTqgJWZB1Hss6nB9ht3qw1evRCj4hsYWtsZ6PsgP7mGtdAuE6eGdoXVZatUANsYjuHhHDrUEiEdXopFq7SUoEG2QrrtGAU61KzUUcq9w87QT8FWSiFxYxS8zVJQHwTUbW9eheYVvdbFGrDu2UDwa9vC1Xk2tUdTHEyb1bhcyxS35EhxJ5iVKwYyT954S55h57A1wvjcTz6z3uJmxAAyd583UK7mu2zWyh1cK14HjsrefU4PhSETBgqjz7WvTLFosxevZNhgtTdZCJQemRRyw2ouAQCK3fatu6q7kvYdCfdfBVXSLbEzMYh8JjNo5mYgyWV6SLFefyR22zw5MMpg4wbi5mdhFEfHDf8LspjN7D6FA6WY36E4kr2CKtE7God8P2gRQNz9bPmScEvVFhhqxhmbCTqKCVs3yeM4GMfj7UnjsjDjLZ3iqeDCeAcYjBxe2Y5zJpBb3nFdiaeHeywiDWBfKGoqCSQmL8rbUqRUqShcpbiE8kDBECn6uuTFhxuXBTo33VctQcWPFWWSfKbLvsWtfa685LbCBYpspqgvYq3sK9NHx7ocLPWGRLD9iKJ4NP8EawQyZH8mgMLhnPeC2kuQ8BJiF6p529y1AdzPU6N1KJ7pg"
KEY3="13horJt6gBxUDDcx2teNjSg4kTMoTcw6Zt1NGKVYJjzrAoPQ2gZqWuFzZuLWGWEcH9rZhineEEkuXxtTvEoHezrJAT1L38BrBSCoaC98YhrWdcf396hBZ1CKH9vw3C7adt3QcjEwka192qnwAbyc3FaDZz7mLH8aqbqwbArQBWytgC4p4BWNEpwLSu8wBhEwbgYA4ySLdF1geeG5qudNVSRwwEyQMxQ4qw2aYMDf3AaXGiq8QT6bG15kddQGAoUt8XLRLTptKs4T19N4771ZHXHRwiWhf58hSMpQJyuzN9uz4z171Dbd6Hzu3Nvae2VQ8NKPCcMVYSwzJmqfogMEeWwqghgWnFK6nsKmWCNXFQHzsp2zE5Ggo8uX3aAXuv8PrNX2B8YuhtVS87WhNKBACvptwGsq7eYzc3PZMA4Gij7oqR3cUdtxZ5X4GdisBXjEp3rHvoy41r8qJujzgKcoqWspeQKfVG7w7wAaANUZytizYJXxyzxRqnhNfQskcgscSgbiAcngh2zqqYXzHGuNSKqxzmwdvjXCygSpzN2EXSHn5fvBDAc68NwqzWhjS626gmmRBcbsf7uk4A8BTMomJq2ZyGWU58tqqo2EedBH386S3G499beAVydMP7HJt8e9T3nSJZMWhKa26gjo6BGitLXSbtMrouhkBt4TwwUJ3WsEZLvSFJpWAsKmR1pZKZsDzLXrzqLvUpq8AHE727yxGKWg7ocKXNPJGU2azBzPqtXAhFeYnK69pkqQGtVXmhYjjTNhPLeiCfJN"
KEY4="1Jh6VbVg3NYSY8ybqgS1XA6vKc8yjaBdJcBticZ542PM5unq77eA8RvDL5n6Buf3bMiAsdcdadT7gKY2eUWbrvBBBgZT49kgPkRLzUQogcfhAWqzToCHZYuFRV3eutSJ2EjZk9juxpjqmAedPuipTET75r28tkS6XZ8r47ZZrpBr8HpGzDxSW8CLdWKRLs3ZRqLtqoesA5HdEuKz3nZ8PbrGboY1mgD4caBjZg838EkqXBZB1BUGMVJookULHmHEwJm4PUMbfZ9PexC1uEnj5Y9g5U6rFAB5EdKUTy7FETzzk8jfA27jTge8xqYyqzjJEGDyZX2pzxa8PHDj32zsD1RxHHVQQDyzdoMfrKsg1X6iLxLxRUXeovysnymXKpt2JUbpUkUcjoRtLqpQZGhCrrLXPWb74SDCt15LAp5JP1wLc5y6soiTN7DnweQHTwGRC2HcWJmJGZa2V59gFiiyg4KfaWZTtqVUPWRMfbrZBRgeKGMBJpS1EBxmN75ZjFLukcbVAPMFx9vHNofs3UxYLkszxZFE2r4H4RyoQ7fmnxgJZFxeoHqyJM3Pt7Mmz9Qjmo4ForFAezqUpiiBcvqujNhCbcqVG7gF2QzaFZht16MWHiV9Y1r3wzHmyhxEq2KU1VXQo92n8PgqBqzuK4sAQUrLj9KZzdELNWvxa7TdNRztLDZ4us3CUgXQ1herQ82y4yp7BNMDXrVGBpb3u6qDgG21uEkzNXBkzUvpCtMmZD53aV1ESkyWdJG9byhbUoWYLVryeSAbUT2XgnNzq"
KEY5="1Cw7ndvY3yMq6vhfuLwxvLKGps5bYP5ueZmXTP2h7H9KxfyZKREMmMpJ6ZUgZFkMnx14QVjZX68mkTZfzk2L4bZyZAv45cP2Eg5kZxmcw6LPRN5kZrN88ixCS539eNZKw3eKWE3yLfoU1DgjKCMtdBMAY69QBb7GPbHe2h2oZXYV73VwzsZjHkyMsv9dKZsEhrqDG9yEx5SBf8poPjubVNFf99ojwYyJm5wpNArXd2KocVQzkCy1WMdjmZF3G7VYHQr4Fz3JmR341dCjWcjzRg98jHaFPX4LMtKybjkVGX3ETX9xm9fL8n2Yx8UWKjXc5Gw1cRQEgU4rPgFFvBhYEgLWSrWbeEopPsAKx9N1s84zS2LRCdVTkar9nsAyYTQQydpdvy8vQEDJNFX2k9zcKDT9ahsqjfdBtB5gLzozrYGR8fPuBF3ANw8wXD6VxvsyNBE8Xps7mLK3qgsZZkmJosbG7sRtBrc2NFutDFdBCL2Lg66fM7fyFSSCsT781eDBiLLSKbAQrLcVbXQ75uVmguMhKaSE4rfZyQXoqESWcxD1gAEmdYXSwA8p7FHn7AUyEgjCBbGdj1ztV2fop1hRGTJLkFZeGd2XP2e7G6Fcxw78Wzp9NdjE3bNZbq6xx9davftowVua1H28hvv5K5DLqzwhyaE94VYaTJwpjrYY23xToykduFJwhFcuTCfP7syMn4NJ5N3AZV1qNDy1gVRi5YTLdUo13xVUDziLYwPQCTgLE1uCWFL66sSHpuxqzD98zD4DD9b1Na6dn"
KEY6="1Cw7ndvY3yMq6vhfuLwxvLKHAKiAnuxXBo3DkdBvFJLPG7Ui2sAkiHa9DgjXmRJ1kActwZ1EX9DbA8QniwJCsE2LphNc3wfM1Are8rPLzQwXio8RZFh1SYqikytSF2D1M9H6dFUVYj7hkHSvXgQRxz3FPdadkT9MjNZtVCrREVXnqodP2DdVJjCi5NpX5mVqYRJbiPQ2eYwsW4NdaUkZYA46ipkukJnBmbC6AzDK1P9diSTbKF51fGcRBL5xi5i7LGbuN7j5mxQREg5CdRfMG4eXPgnBnNo5Tw1qv68owK6fCdPVmvBdWvZHRcfKYwepuMJ9a8HbiiDwChw6JNTpkwfjLT7AhtPa3kCEheBEa7su2w67dcv4ncdRXEnTp9cC7cYvu49Ah8HPvkNcda7E72izGLPjBhA1gWARp1v1ndgtUTeLmYJSmow8xviZ1GHJQaREiWxRYy6RztgWA464cgC2uGjQHwmAXbL6azM22R8UdsSkA81jxqRXpBD3uGHdRaYsaVegkSwpXU3DQop84ktmKG93PoQ8SVPrfR9kzdeSoho1LZ1TwLQo9nxdzccJELMUcztnmFqTwBkiQSMeRg5uoALnxwaqF4aDHzCUzLQbhxM15CWM8jMnNdScB2wfPH5mxEQq16K5fYAZXMZ9X4DFuGFYmTCcShDcGv65NbPbaAihRuHNSZqEknFV4xBmNXdyBC5BMXgeZ1zddeFiiqysq77h7bVhRLCdxNinv8X4pSWoWyYRMYJyF1pSPQr94cQHYrNd1zzUj"
KEY7="1ufrxRyJ47kwKA2Wi33bpQG6aR4r6eLYXDLLfyaFAiAEkndY8mJGzwWyWvnW1pUpvb4pMtwGDuACoJjP66Xt983PrhB9xBQGa8hEnHRkH7fqmDLPoK6nzbruAtst7vDmR3yHYPVJZGDQaqNqkazeczhnoRZsf41zyyoZY1Jt7T5jTkfh5Hw5spbHnkrhGJffYPtQkCq8nJGqTPeSqJntzFbFriokY1paks1Qu5jK8ahVNj2zTRfoNkvuUEJ34PeMQUibbE2BkgnMEPGqyLb7P6J6hj6hmvMpR3oXdi19uVpzdDTaBz6nrRhCFMB3YiBrEhUwDEaMfEB5C2HZpnZs4CMDeh44UUyFwHcu79i3FQthrVwzeZ7NhRA8nUBBoHtd8BF3Bd3yUdq3TZ7RiqCNbmz4XNgC2C4R5esSQTYHfsNTpduDGoaeN4SfayYHmoqsMXLM4XEJatK84xgFXtedS74sbwpptz1mhtj5R5s3oW3fJvdDCamJPvwtNepw4FJ4vTsveC6iW9LoZExpmQSmRWf2jiMiszMnaXt2a3AGkiw8rtFzqspsBWcUYbKKzJYMBf3Hj7wCZ4NLikEhnj639Ry6Yqd3BDVwwmzp7Dbat7hnaMQbQEUgXGAwhUWaPmBZFhH1Rb1bZXU7WkEdbjaWKCFkhQPfTmneh9qRi63NtyT26m13JaYpy8QAbjCxJoa6yUftE2Dqh413u3V8zECajq3xsKM8mHhaMdz3AQ1XVT5EhFchi7dKwJ5nb7TJPvz6sh91N8xagjfyyU"
KEY8="15PSrUxfbGZyfBcCUb4hhZhPSXZVCjPYAj5btNanxYYHC7UYBy7bycbXRwwkyXDouc21CpwkSP7GzTM99YH6AoG28R9amWBvbiuA16UheqRzEsqVpSW3wZthRokezz7wdUrRpNVuxqhDA8Ui4EfnRxkGxKP1UFDzQu1E7FveEzujuzwmDWFayAHVkTrreSsxW76c4uou7T75RNN6JBZycV2HBnEEftTbRgiyMaid4syGraReUNyMSF1U52C27WRkPSSX9dNFo5XG84yKo6BCwaZBH9G6BBEaaQAtSeHgFifo3kHUBcMNcSTHrYPBc6sbfDC75vzx8Cr8UNQ6rdV6oPVAgXp437PQjx8WbgtJsY6EFf4TAR38ufxiNZRn82zZbvctQWXPmWQqqecrCCM6boS1c4QHyzxiQJJGu8KfuSGPEKxuzSeixrAkymEVWVDH98fB1a73s5ZAGsnUfPSoNZMDcqTe6PUieXjH4QfTSX4re6PDDEPm3t7Pd2cLnDKMTSQhNUUhzDSYwjyjzz1mYuuo167Ut5pc9JNESv6o1hfaz8ZNqzHxzTAtMrcz2eJZzApEbePiqX5ws4EefDcWZmo6xa2epJT4HatSrGr6iKuHk9RxcqWnF5GC4bcdL9muMzhz5K8PQ9t97dcdikwfyJpdFGNxAGtdoJ4xTCE7ma1QoMQcQSYhBH5D6mswzHYAqkd5xzWSWRd8zuWMid9QMN8ydSbXLJnV77VbCg1jkEmF44umdvAjcmtbJ8gTtNW8mB75G2xkp2MnyJDixXNVr1bKYgq8rN6"
KEY9="193qwaJf1BefwzGYyZCPkVEW6ub8KoPLxT3hdEpVfduARSqMvnrMukhMgQYpxeV6Vk5ZM4QQgjiUmGAaACRZueB9jgCpnhwr93G83h56qWk28DQ1v76sN7cN7KEstiFYhDa8tmtEopCqMFJGALTzkzuNfTiLxUYdFaCkzbmQThz1YQhfwimk2tKjDDaFndqWBeTZkTkSdF9TvEw1UivEGJL8UHFm7P381yByw1y1qYm2HGQuE7kyzm5Pxy3cPujrv91qJrvxekVXFdS2jew9Dt8i3V3T2K4yH6hm7D6v6zg89WwjCfVGypgndX25X8HBghjNBD9nwJCq63rShL2uT7smmCWohxai8BevbyQ6t6gAVHtQsHrBktWPWHqcXqsxMRibyJDkC46RwRxM9B75foYDQnGdCVpcMHQs8tsEBGjfBgknAcJginzRa8Rng3uWUNRsqXHPoPuAcBAjQdnZTja13yYRm9ZVcbF7D6Nt23G8Jmc1kZzNt1bxm7PSAMKRdGAQHHLQEcE92qAAkj7CSPByFrmHqFy7UYXg45eiqr8hknpEyPDzN5t5FTcVBzvpRTAMvhEeNYUHBDfG8v8t4JqKYXAmEJspXB7HhW1pG33zTnGm8AWsuxKco7qZXBbnCzTdMybRSiEcGZRLEZc6faqLfEfkuQBh3hfaNsYhaCyu5izyenLRg8Nsnz4WjEHN8MiL1BX9X6ADzwWds7vHZfs1FeCtCE8P2FdazcHpwQDeCDVH3Kaw4sd2nSDwhA9ywuvB3WUCY"
KEY10="1q85TGAnTgBBSch1CSkNHMq5moG2FFRWpxgbdhSzWYwrA7mitK4LqXHegypZzEd6Ds48ZtzjCzbCnMyXCBZguKFxLVPzMD8NiaPmZGhRHFKDs2qoMmrggGWoSYWFfF8yuRY479YdfCtwgC3gixZbUHgxMgK4xYLtZtnVdoBpkD97aAHoQd3UxeUXzK54stN3wWniMvuYx7U37ugcXP2t9zAdBvFB6pWK4HhnZt6ZYPBQpdp7eLtTmMkvWAjmub4Xv3VGLqiuNaQ31iNfFE2kZ9W56jD3gFQngU14ibdtLEuAKwRKvEDWv7aEh5BqQMXJ7SjtfwNxRwemVJNutiH7qDZWdDH5SDZb93huv9BXBTQov9DdK5UBChvuC39eyQiTiJ4qVwcyeusMBcXiSXC3v2F1WL7xsjj3XeUcEnp6HsqmRgZveob7gC155KNmQTsKSrn6cFqfHDmYFNzvgwVztWJaDf8L6tELSR1i74VZrTXypd4ju85fgz3StZ3bgdV1XEmN3cXAEcwNRHFtfxpsVGbdR8AHpMmmDchz26iwSYu1xGCDzSNhRcZz8cxS8Q8UFxX4C1a11ayWKrkq6f11itRezVsYHJ32JJkDKzCDpx7vNCPbQSoBecrCqH7mfGzfyTPzSKNTS92tuW6te9pyLejdortnEnz4yGPtnu7kgNjAmdVEBGfmn29i1iQfCQvbG7T48gDJJ9QzcH1mUmw91WiUUXJpea5dMt3wQTsWEVug4PAYqAATZnf8wLM7xvu"
KEY11="193qwaJf1BefwzGYyZCPkVETgkUN1RouBriJQBin2m3JWw32cDAqY1TbKBSF4ajvCGvzUmRjwQ4NNMfSZkWETMVcuoBN6GtA658tYXFvhmMFyUcJrFaeyGZQHnivnY7X4xZTE3yTb6Wzxajh5pwhnAjMbPykgVQXB9fsRZoV7MkYsLrRBMf5pb3BYTExp2CeqJ3DT1mxQFWB6Y8YLHuUuNdVs5SxHnqegrbzq2JjJ4uf2kGFiQnBkCGDEKVBuAdbPrLzpwQPfdYeivqj21vDJEJenRNBfAxtnT4yayvZgdUKavGR31KJMiJ4zhkD2b5g73A3xyaKJ1aV2SSe4mQCrWikCo83KD9kCLB6LazTfrFmbhmmosJERz35bwwVPAihoSD6drbxfgNvHCceLy1xv1GQL82isx7k5Y44NgLKfRxhMqr4rfJVb8Rb63GoJ247mKrq9jnTxvzR4vF1hp5iKyuz81Sx1jAfwQiR61di8KzrFKVeczn11oQf1eCuRezsoHR9dScUbZPy15xW1xxM7u5V9hDbtKn2embUx1ZLXAXpGBgwfGEoJNmBULDTpL4tBa5M53EXMkbbGdSuMex8KJAFPGS39zV5ahCRVibJqZ7ag7kEnpR3qyPufEVXxfxHMW1V9isogCe5enTKpnXhisKs8n4ufD63ztxDvFe9pUgLmbVNNdNWxhSYAknCwCohfJnL2EcVtbo5SdcX2krAyyXcxmC8UUf2Rkuuo2wJq8MVzB1gV2p2MUyjnSNVSD2TfLvxhjCxT"
KEY12="1Cw7ndvY3yMq6vhfuLwxvLKNqV26S4uXzegddqmpypGASd9gR57nkiDcnApcvna3EDncfM3sTaCF5CznKQ5PditSEa1y1vcH7mApDVmbtDYEbhvA3kvFYgsuftJRYQ9pNHeQsuxCJLJbD2YJG8p93zPgpWgYpK9nfMeaDyA4vqsptb7XktTupGohydr1F3Aq3FyucHxntMEGEW9GuNHTTVfeEK2Q9d78tPL36sebmmMammvPbxxHcYEu3Grcw1MiKjsixxggGDkthjCfPUXGw5TLZDCrnZGY5YELC5vYk7hZkgrYZsoMj17KMh7udAVvtzUrojUwxdWGaUywm1YsmEisLtjzr9wVnaxqpifqruXTwzUfgxVHPq6WJ9ssTPgdEeWPZc4cxT2riU2GCpLgadjqdz21ZyohwFv1EtRUkic3ghGuyerZL1boRYfm1ZM8ERPj8RfLmVY8rKxuVtppe3Lh2uA75XdtJnXiV7XM1yX2drqGan2ARPeXnnXiuXE69VsbGkqhKEdEA8qovd5bz78vbT3p1WGzQwbJgbgfgzKCrEim6FwcXMHSJ8Jk4d7ZuZpJrBDqRqnvsbrWbVpH4ZVmfs2dPaYtsRoPST796v8XKJH8XLjW2hbrxfd9zBNZCiQuACvadE8634SdW8qZd7j64C5c8Y9YwfUxZNBwry4DemvS8F2c73pQx7V8go8MfmhafqWvPKF5CukeaDnxEX1Ybpvctxm8fPQPei8CzeVDLHxgxgANWwYwdEZLX8ZM1kfGPgma24Dok"
KEY13="1ufrxRyJ47kwKA2Wi33bpQGCFvRaiwLneuAQYu9GhonakLBiJJobc75xYkdkUWYPyVbgvpUnc3xBgGkikU9wo2xyfoxTZ6EQD69eduqzojXwVfCsC9Az6uF7XWoTQVFgs8eELF8ShXBjmrp3RZnkLz3axATacwtSX2q2GdEw1NxehNNckq8rG8NURUmjT9mmsbrKAuyZd2xZ3G73MYyD9siu3ULyALAR4aTxVTEC4ASK6bvZ5ynp5UHcNN2XSgtZdQB1hwaEk9XLZL1QJvXzWC764L7URv2kaqn1EYnU4uigMeEXuPMyiVKHGhbTpnWrW7CzgJKJM1PJ4TSDWsFfkA1RgpCvGLV3tq8xc19vKbL9PtemiUL2VGj4xQt7F9iFSXGf5WAuwkwANGgNZhNTTM26dzUKf24hfSSHyRrVDjAXKsMziVjktSwCn2v7zrRbJXQ6aj42fU29SRCdTtqbwJaHgo43qUMt7SAeYbE4Gs9tT6p9XUjJ9aUesy2vLRFSWd68fZhQLqNTn6Bm9gR9x6RhcaRehdX3eXfKZNsr7CxkNofX87Ua9s1mdfbQ7LUutKKVnd5dyhPqjzC4Xxb5iowmNSHggbr1uGdtRKVPdB1zXDA8VJx14E5JVGvtu3n79yCNAHp6cQdJWJxTF1vsVPQXHFB1ffDo65gX9DPBMHQZ2VmGCYnacw5UfNFxH3A9USR4wBgxaU6zWktkfkUXVcFfS5kXHiCCk4wFXE1qsg5iHAbvvGxvwHfoT9oXh8Kv3NgfWWSuRo7r8p"
KEY14="16wewFj5FeWY4MtFBs1sXQWCh96ThbTUuykKCLxmphQRh8sfKS9pCYx2wVVXNHkapRpua23Pd4pi8NtkynB8TJd2iB5gXtU2cPkj2ebfmUCAmWPYsPy7oqg3vyeUinF1vegXV7ujUfUGxAmixKYo5JnsRAdC3k6aAMBT3gMU6mN7CHq6m21XGWgDf8BUxSiJcuTsUpiJ6gFKAoFXHdAZSjiAvNiEd6rWJP23KCDhxuQyZEmMNStjWVDcGffJA49er162QdZTmxXkB4YphpftJkRhaRngFMqXJk6d2c8L4vwNzMmk3aveZreUFn8r83B2J6oSMHuw5b7HkNGqGs224bsa9fZgNRwkghgrTPBmkNm86mz3E4etVkd7wtruUeSaiJq5YAb1qFRkRGdCMNev3bfFgbon5VX4JCbLH13MeBR3AseUrky5FmoZ5VBmjvXr73yDqd2315LSmsEBeQfjrie3FBGpXvokgDvb7vK2HKM9vop5npDfxmfsPseyvXi8Fiija15ad4JxDYoSeKdczeHrocwt33WqCS774rzfaMnN7iDPK1YL4HyrEK4ueP1cnjtj1rvRvy6dVRSJHtn72LFgqgVZBHsbw7YQ2YHiovvoEyuv85CczrDyKYYdry5gQ89Ug5WZgHPQ6DwdKDp9vpYwtM8uoLe535noa9rKbt2DhkQFMQ2gqaopdSJFgATNmcuJzFjWqz2sPuqY5pneVRTyBNcX2yhoEasJvqiD3wf1wEb8ZLSADruAXWngdN7iegR2KzHqGqTY1R1E36Zn"
KEY15="16wewFj5FeWY4MtFBs1sXQWAeojAgZ8htZJzU5YrXFECfeGS471WkH8YEH9gmRwZAFJ5BpB88PR6UEGPpmQTdnn5bx8kkWkJG4KNGyuowxYjW54T1C6Mu9CbvNnaKUGQrEFD74Qz1CALh37KRqvWQ3NjT1jRji9gVfj1JiW81iVVyC3hrYMDT53scuzSdeQyog9Q95RtEPbo5U6apvF2myAcqpwM2ottvW1pHeXKy32feQXx4Cxjc71JKCxnjBwotT9uhiU722cp44XWn4jNvGgnzLKKzHQa9PFcLszNdvqjLJKzMGHdTVfP5GKG8U7QNEgTaozazMmBoB6KWkrbkUK3XFDs5NXtjBx3DY68E2TQ8MuGmqqhpFiFS56oGbLzX2gsUpduTeXBB4gfh2MaSMtvFWqC8ZpSv2krGdr8HY2zc2gSpizWQppAaG1SwWNQd1XzDfit2tBKed6X2M1aPbAEvoU2eCtGj3vHw8kcYmwvi4WSHbHtbss3SShrY7b4MrkCYj74M1z8QqQS1iwpuyBmdVWoy1z7rHNXUzuPMQ4bFvucr8kYMAR14LKQDmT8cBrdMHXMEXXaQBBxDLxmr7S34aRyk2nZXbidsBCFPAhS6bHpEGmogouJCyBuDtvQtbYhzTaUgAEqxvwZ3fLawNebMCUGtjSbB8Kaw3eExcayebqVkxvKKLDrdtBB9CMDBheYXbjj8m97vwqedpqfsiH7YNDLRHsn9jPC8CLjGqYcdSRGT3CTiknVqfgEGzeWNSGN7hN9zhnv9H6o2Uop"
KEY16="12ELu3Qi2UhYTK7sNHJq5rjRsouvXAgNo2dV26wBPK7QaveDLup8RbJsFTvmEW4BfB9STXpuFMFjdaKHXtEjUfd8R7HdAwcQtGVvw4JN9r1XzJn91EUn6xDoAAvecXvcEg3yLirtcdSFJpHTbpfG1PH7NoVF3Ufdx3UFgxfMwsKrLASiAxJgV6pt7Mkfhe7yasyNTbeGEtxPDcLvowMLdyCA85zEhG3FuFF6HjUbys4LDAj6yYfufaDbU18xRxUG8jppniNZLQEcKBjnq1yNR16Uvk7W7t2Vc8bpxkKNBVkipT9QLkg3WjCoTzV3Jm4Qhsjbia9U75LHr3iHuSv8B8tTsJvvwyoGZQwJ8vwQACNoE2g2gus6w97wWW7FA44HMKDdLMGzxCHucJusnETsFno5ejjf9NeyoWAXaV3e7bu4yGaR7UaTAj1GqvS9AG1MveT4eBRpNUvdyhPnkfKP21snLzNp6Dw5WB5M5wm2XS5mSPGVZvHD7dRZN5V7FtXAHRqWzwVRib1wSJiMmQbv7bh1mJiKzEMadTykLYR6LWhhkWKEh7wyNTT7QVHJFdhnAKx8jzBaTh6A8vfqsVPTzQ8C7doykhtWYpTXWJv9Q1Eiy57Kb8jHyycbWTAKyo73rEdNRNCAL9ZiSsYyccnb7XYUk7caJ1e6aKcTcuGwJDRPkUioCX2Rg83KTh7uE717evWuPwD1BF2FR2dxvM7YBeoViKvqPshrkoNmKg2n62xia8NQpuTjaSaHXyaqZS54DoHb"
KEY17="1cXZ1szvjp1Guhme8d7HQwcihfHV2V9Tc2SM7oxQ9CtYne88iJywgEpWtAiRnYVAGEoHLuTrW1z7ccCye2sVUxQnTrqwBJdvSch87WZyDFw6YpjEJjRhS46cjv8NSTkFUVqdYMmqC45NDHbqzZh7VKro4Mf35C6qaDVAJ9wVgPdd6J4SKT3KozMKEyJoD81CChtsodQsXMTkfdvwHyPjtSEkULVBheXQuebRcGJ7TBdocMcPeXaYa79sX5vVDhaWu9FW3nuTowjks97noVLY2WJY2aTePZNDhrfLR9VcfxfvXvbxXFuopAAeJ8TF5rUdnSnu2RHXyViUqxTPQnqpM5xqAfqMSx8C2drotH2LGVHQuPXLkYyrkVCsmVK8snK4CFpLkVNTqP572HkZWzAQy9vgX4ajvLgmstF7GV4ai7n3658VfT4GGeFD544UuoqsNdx47b4CxvEibFtCRL94EX7DHg1493WLmM8dyDaZX8QWaWNYpybQx2za5GjAzsZrH94WRrDYFn28fmtA3VSYMzAWKgSMgVWF3zzEysoTmvoWSVz1V4eshaUPMdpiKDczdtciH2v8HnNsuBqEzjsyU9ckrfPzcMSDveywYYbNoAoWZmGBpxMGCRX3H3Nt8bZKfDGJVtAiphpxmPVsxrAQMrhaA1P95b8VYYUQZ8D4SyLDSBRunv1J8Pwni8qZMNrc5P9oCwnoHZRV38QiBMn2KW3t4kkAUaSGZyLvS6wo7jtNnCF6TGCzExrmWfQ2YQjhh7bpB49PvB"
KEY18="1BuMVcA1nU13NDAWxA1JxpK2zT7eXxpKNz35hKXrjZtQ8qDwx7gUYpnZtk7jozocE74WbxpFN4wqFDEK9bGgoK1vfNq6q5uv1zDhyxCfuKBzjCa4yGxmA3HeZb73derUu22bHMeTVdKoMLcfxSghC3poaCAFtqCCo791Hw4fJUN8bfEdENfNzagr8SZmHCEdjcJA3RDPrK5jTTHo6NkahVtjwKy1aBNbF2tAmCD8cLxq6kKztuKJbkXGhZyDgGV5rGeTo7p4hyHhytPVYnVsjXUymSbe6gk8vqUHkonB9d2L7eL5wG6zcEWdSRhBBFBzkqhrnjd3fZ6gnPgK1q86FpXRSZZHbpntSMkY43GQ1ujt8mRaK3piVdMjCWsuGMeAuSXHb6J7kDc16FXmoUUhLqoaU4jmVpCAJfYEgQJngv8H3QDPyQZVRnYVWhxEPjGWNVmTFR5vgR13G2GHP8bpv7tMpF5MESexfjYsXFJ2EVe16e3vi5eMyj987Py78Q1kdTyCtPs3PQG8wwy41nWmZzz2JfHqmZUKU9DfjNfsAvKehheCzn3SPNzvrgxsBjm2ZGD91JaUQLpDnGaFte6BGtBxuMSMKfGTuYa4QKwUZ4aDCmhoPKd6G5ftKUFC8npxLLDk2V6QsMCtvKym5S6QwciJiXnb9iFbwefCShHoJATZcxsj4WhBnztcQ94zwVHRTAMkTtikTzmgSvGzKzyhpeoCDo74phXmjCLaNNGSm613RY4wL2P8Xytpz69F32"
KEY19="1ufrxRyJ47kwKA2Wi33bpQG8JXi8qJJvTxFjST2LqoRWuDtB6zukDSXusiNtoVAghuuUMgGvzZTB5facSo6e6xN66gm5n5aJjz62i9vtGkj5uo5VZ4gwHdqzGWaoVdDZfDLfgeZq8eA7K3Mk76hCFmbtZxxNUSEpRGF9E5w3UQPfDN2QSzHhKzXtvWKBa1Ka63oKKctxjhx5GtnAUeazSNUuDvNiM72BJmGBJ348FxMsteEBuMjpDWhRBCZS3QMPnqkhJMGLxQMhnnQojZGP4KXWUjLCSBksDzNN2HCuzdVoRVJsbDuV3mTgE9eMwnvY2Ve586Gu3ffsMwSAZcqrFMzrNdRgK7jHN9ZtiwhTKiPKmvZJ4mbYSfjxWBwA67w2c3H7GUr89fWo85ARnrGQBV9z9Zxugth27GwB8JZv6vFBp69rFm4Wn2vwRt7BEU2EVD6hJ26NS6Ra6uu7w9DP6xFSbWti4PbbYMp5AxeWK6E9Yg1Us9oSngVxFytuqcsHbmFFqhRDW3Ux9VzG2k3Z1CRNSGxYaQmM5UM28FfhP528woQX2bdRG6WiuKTvwhBHB8LDKggRTSN1NvrCKRe3GjrAdSbT2ysFUBKLH6DrSW58XZTqdMSLmQncUieger5rcp5XWTSx6QvP7P3TPWY8XWrh7MCNFiQE2L9zemWnb9ap3PXM5GFCNzmCDk8upCA32Zy6pC4EWg1fQdYfiaEh5ihvnQehbpymgM6gH3trqZbwDnMYnRU3ydsJRug93X59Nzx6torGdguXtR"
KEY20="1cXZ1szvjp1Guhme8d7HQwcauz2FzuDGos5FXdTvKWCoTWWGfRSdDuYpuY8GZmczoSGqYb5eeeoGEuofWa7D2jK1yUKxDRr9VdGcF1mWkmAqXge6vH6gCnMUxeqbWwNtCnp8txpWbehwqqzuakWuheHZt9m2zep4w6hZwkcavGLiW9XR2mwfux9v9Gzths4W8JuGkzadz74ChdrX5T2xFDbSs8qLVULdgzmNp4iNYu8oKa1A9Jx3gFtpbB8aMZKmk3iZgVdovcL7pTf5DiykfXUjFAiJMNR7mP1eDfQsPey3nJzj4R5BCv21P96V1ctuFk9p7DvNaJA9yWh5mfugS36f4SKGcTf66DF37Nyo2uhiaVyZqtbCnuFB65HDXa8MLVWK2rsN75cioFV2Qt4fpChircdH7W5yrzKhUzsQiGggsLJusNWzwUps4DBy3RaoiZJ7e7aRHDqsS2CZ4xwPgfL3sRV8q4J5fjU4SpdXpbKYRujR2jq3fT1cv3mpRudQNikn8DjiQVC2PoxeSFjtdz6izJKSSnPcALgAhC1mVjEAhEFyF5mkwWFXNtSuUNwpKG3egv1LnEF3ZjjgECpHjpVTZGUXzeQ1nD6TUuT86n543PEmwX5NDke6rzNDQ94CQ4N2QkhQmeWP6WTk2zDxHeDHUJw74UowESPrppJWPN8GRzz1LpEVkefkhQscvaE93YqiH1hNvLRF1BZ8w8W8jAJYZ7y4a6jGo8pStE2jw6wG8P8no8LEzDwk2ar46vD13Y52kUsaY8"

mkdir -p ./data/node-$1
rm -rf ./cash-$1
go build
mv ./cash ./cash-$1
PORT=$((9430 + $1))
eval KEY=\${KEY$1}
if [ $1 != 1 ]
then
    ./cash-$1 --listen "127.0.0.1:$PORT" --norpc --datadir "data/node-$1" --generate --sealerkeyset "$KEY" --connect "/ip4/127.0.0.1/tcp/9431/ipfs/QmZ2UwF3fqrpc2eAtJbgQytkbN2UnLQod5LMfswU7YGQDw"
else
    ./cash-$1 --listen "127.0.0.1:$PORT" --norpc --datadir "data/node-$1" --generate --sealerkeyset "$KEY"
fi