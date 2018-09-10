if [ -z "$1" ]
then
    echo "Please enter How many node(s) to start"
    exit 0
fi

TOTAL=$1
SRC=$(pwd)

KEY1="CAESYKPGwbdMn73bb1dVoNDliMTrEy9ui/+uu7IqrR5xABDt1PgXkrFB+D4v6hrrQnQqYlwNkzD/R9jhmeWnGq0IPyjU+BeSsUH4Pi/qGutCdCpiXA2TMP9H2OGZ5acarQg/KA=="
KEY2="CAESYMTVh+plO0Jk34qWb1vc9VloYfxyTIHEtsEKCxMMvuVXpOR9Lgf1C5BwU5PbBgmPgGRQR5ro7c8LKE3IKrjpLDak5H0uB/ULkHBTk9sGCY+AZFBHmujtzwsoTcgquOksNg=="
KEY3="CAESYJHWT+jDxKqyl2dyx2G6m5eg606+ryEXz/a31X6OmNcEyBM05Qk1h6oYE0K+voqhT1cEx6egDoi/cCIhO/lYpufIEzTlCTWHqhgTQr6+iqFPVwTHp6AOiL9wIiE7+Vim5w=="
KEY4="CAESYNiZZduhN5gujiobPJOiOOMsxOWmNf0bC/k5LAaznv133NV2KJ7GE1i2bZyc2gYyTtEdlYlhjDMrrLlwoSGlKFrc1XYonsYTWLZtnJzaBjJO0R2ViWGMMyusuXChIaUoWg=="
KEY5="CAESYFS//5W8cbW4Cncpd1L0dXSAVLC3xpa7UkM9o0zGSBYWnF2wbgvmBR5H72F025kB7N3AJLYlJfCra61h2LqLflScXbBuC+YFHkfvYXTbmQHs3cAktiUl8KtrrWHYuot+VA=="
KEY6="CAESYP+D/dYGsG4y83uqX+QgJtbtbsBhwjHjDPv4ow3kTNsBS8C23VYNZJg7+qJvBfLNfvdvD/h5HhFp5e52QwdbuKBLwLbdVg1kmDv6om8F8s1+928P+HkeEWnl7nZDB1u4oA=="
KEY7="CAESYDonwemIfmVTmlgL4kArbZiDp6uzkJOdfQjGs7qe0qvhORgKj4rEDPFo5fkR6zgdpUhLKFLYm1SlA48lFYGdQUI5GAqPisQM8Wjl+RHrOB2lSEsoUtibVKUDjyUVgZ1BQg=="
KEY8="CAESYEzN52C7t9xG/TOh7SCiVrol1IyW8QrwY2GM1NuK4n+Asee3/I2rTOYL3i5QdiksRVkS8TkFtKJBna6G/OiC+Y+x57f8jatM5gveLlB2KSxFWRLxOQW0okGdrob86IL5jw=="
KEY9="CAESYCasN+1QAyT/kLcKrY2Hi+OTR1+oyuFsDgJdb3lCPUvJFBQIjU40T4/3JdqwCft1652SiMDYB2YrX2nGRz9kvM8UFAiNTjRPj/cl2rAJ+3XrnZKIwNgHZitfacZHP2S8zw=="
KEY10="CAESYEzHeLV1HoGUFghx34hEpp46LDm9cpcH+qkniY7quJ3T8IIN6tbb8tOLyHYHqa0AWVulrloH5sonQABzdx9hGGnwgg3q1tvy04vIdgeprQBZW6WuWgfmyidAAHN3H2EYaQ=="
KEY11="CAESYB54Nblh8SDAmG6cza1UuVW2RTFqvHX/aAVJsH4sHpYAHR3i1guLSE07QV3lumiTKx5TCNSMKhJS+PZ12eqsZJ0dHeLWC4tITTtBXeW6aJMrHlMI1IwqElL49nXZ6qxknQ=="
KEY12="CAESYECqKozc4eRnr2q4CemkDOUO5LyHb2gNNiwbLkSXtIc2fGb1Zz4ovAwhVaLwlqy5b/F+qVFLJrdNerMeqHdW2+p8ZvVnPii8DCFVovCWrLlv8X6pUUsmt016sx6od1bb6g=="
KEY13="CAESYJMlihlX4UjQH7nKJbJc5KyivfEudAxMFR7ObMWZopky39w1R5/etdHs4H3OYnaepEyAcO+SP5OIPYXjQNZfWLbf3DVHn9610ezgfc5idp6kTIBw75I/k4g9heNA1l9Ytg=="
KEY14="CAESYNJhho5SD0SELq3I3wh0JW97RSTiZO8oNJB5DxKF/1jqgIn48S2ASOpUfyQHfhPaZtQ3iZtEKUshRtDR02rDXiGAifjxLYBI6lR/JAd+E9pm1DeJm0QpSyFG0NHTasNeIQ=="
KEY15="CAESYOUl93I3Zg5lt7+WmgyU1lgYLcz4MkomNmdVKVK2F7ImasBUSpgNkVr/mgA0O2wlH9UV55BpNClwWXbyzvVzjUNqwFRKmA2RWv+aADQ7bCUf1RXnkGk0KXBZdvLO9XONQw=="
KEY16="CAESYE/Mkpyl04opferid/a+5RibY/LX3PU8J2JPM8MllGXZK8t25TiyU5iyLgONflE3Xim/CNiZdmPgTDX6T3aiWVcry3blOLJTmLIuA41+UTdeKb8I2Jl2Y+BMNfpPdqJZVw=="
KEY17="CAESYFLKv0xyJ/UAGZcT70QYGfT0A73NdtHLjjvG6KGU0FDMpUXsX4EeJt1YQHh95CQa2CWbSMvvDikw6HArXpwCDfilRexfgR4m3VhAeH3kJBrYJZtIy+8OKTDocCtenAIN+A=="
KEY18="CAESYP/COQcsbsM0sP3y0XYPKc2n0zn1nk3vdpouOwaWspXOXHLRHXVvA/sPrLdNoPOnJ/QgWMePFG7M9UPNotNCQJ1cctEddW8D+w+st02g86cn9CBYx48Ubsz1Q82i00JAnQ=="
KEY19="CAESYI2bqZ1HQ0F1EcKJ5mVu/NxCCz54Vec0mYZQg8U6UvmHCSkHNf5HBAN623QBPZ8wiw11jE58JZq22I9i4vstFWkJKQc1/kcEA3rbdAE9nzCLDXWMTnwlmrbYj2Li+y0VaQ=="
KEY20="CAESYJefATvcGybItEwPo+x7xJ5dXE6Uw7lPpMTlsYrSJLxdC4bCGimi3M3FxnWKkEkiFPu9bc15Vqj/fO6Wk6Fwnx4LhsIaKaLczcXGdYqQSSIU+71tzXlWqP987paToXCfHg=="

if [ ! -f ./cash-prototype ]
then
    go build
    echo "Build cash-prototype success!"
fi

if [ ! -f ./bootnode/bootnode ]
then
    cd ./bootnode
    go build
    cd ../
    echo "Build bootnode success!"
fi

tmux new -d -s cash-prototype
tmux new-window -d -n bootnode

tmux send-keys -t cash-prototype:0.0 "cd $SRC && cd bootnode && ./bootnode" ENTER

for ((i=1;i<=$TOTAL;i++));
do
    PORT=$((2333 + $i))
    eval KEY=\${KEY$i}
    
    # open new window in tmux
    tmux new-window -d -n node$1 

    # remove data folder
    rm -rf data$i

    # build options to start node
    opts="--listen 127.0.0.1:$PORT --datadir data$i --sealerprvkey $KEY"
    if [ $i != 1 ]
    then
        opts="--norpc --listen 127.0.0.1:$PORT --datadir data$i --sealerprvkey $KEY"
    fi
    # send command to node window
    tmux send-keys -t cash-prototype:$i.0 "cd $SRC && ./cash-prototype $opts" ENTER
    echo "Start node with port $PORT, key $KEY and options $opts success"
done

