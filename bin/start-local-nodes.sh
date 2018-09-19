if [ -z "$1" ]
then
    echo "Please enter How many node(s) to start"
    exit 0
fi

TOTAL=$1
SRC=$(pwd)

KEY1="wcCbvu6lZ0BESklTDOQxV8Uxo0fLTQTUJWC3rKBU1T5uJuobiqEcFK+haNMMnhJRIezFHXKvMvkKrrcJUCiR8A=="
KEY2="6j3HrTIjzZqJpCsot3GajXsMnYBteNiM42n0v/iMTpWA5Q1Gcsh9muTbvvMDjqBgL32IznLRTjLR14JbDJwVqA=="
KEY3="kmyFT1QWJfP39kapBRiBsbbnBpLI86w8Lugu8sXTwQIK4iwyeWzbpOHxNqM6Kaw2z7JMZrC1v/elVAzXHG9eAQ=="
KEY4="+VGRuqhYkX78c07pd+HL8pcL08w/6C6a/1xDCxKKOOlKKL/mX+tNpsocT2uP02cDJI2qunVHNNr24U/Z0YbyjA=="
KEY5="McY8Wtv9jiKGLB0tmAXArDHlRSxdCMhaCs6T7CChXs+AeDqM8UKt0YbmSWdtXcsDi60Mg0mVIn0O6u/LAB0ZlQ=="
KEY6="xWK41uhg6/TVyufpsR2Hpxj5pYvOR1zhWfgPQTv8uk4BtCKL0YrAMlJ7qz3D8yW/aHhWQaa3BncI8kGPDF1bgA=="
KEY7="FrS3RpZ21MatJp0+/ahx036yxa2pBpvKf6VB92zOv5sMAyeSpV3Q9ftD4TmvajAZCUZWN2TRsb8zzt9+qe9QsA=="
KEY8="KHQDBlFuEagqgc9WN4ptX6Y9TeHl7/mqQiFX6rW/owUqliUp3tSaxn2JoT+pjq7/syaEIQBbcJWgdspHZmaFYw=="
KEY9="fApYD9/yQK3rpGN5NB05fayxfJgP9p0KOsRnp0iArdLLWgWSSy/2og/F16iauhTfVXGpJ8p7Uat73uTQFI5BlQ=="
KEY10="fftkt8nMG6trnZKtkaNEnuYxVJMrZA97Q7haP8gJ3xFpmwk2QGWncg4IGNLi3v2tJ6iOWyecwXJ4AlCC0aFaSg=="
KEY11="5TcK7QTvzBsrOP1MwjcytDbUZS4kbcuM4EKBS09us3FFQECyCmDGlDDXMANblmfOolA9q0OBpds1GbU00z3ZKw=="
KEY12="PSvoYGAaLdazA3efE2AoEBZvUx+blwSKEobnGG3WFxdyUBXdftK4NLnxmS53h/JISxqP4FJ0Hw/1J4e3k3kBjw=="
KEY13="IZuV2roGZ1EX3tmb+2A7LZYB3L7fv8ccVzfM1mkceh8bNikjxc9QeYj7TN+Ja59uf+GKWl3YMTP5af2f5Hn/hw=="
KEY14="4+k+SKihQQEHr5SMPVTt22L5FJv+fuyBETlPofChmfb0kflOPCsvUh2ApRrVrifYqqQJuWvwWDvmpRQMH6FtiA=="
KEY15="8DW7wTtHMRxhqkNWIdKsSiydKWhLyAgA+3pmUJTaTZkAURiUPeicerQ9Skqr/UhvgBi9pvNoIJJr97JUuqoldw=="
KEY16="4HYOHBQSIipfJ+iagf8FZIzTMtesLd3O5jf9hUe1zfNmgTUxtHhAWroeEIDeeuAhs5IsdVkETMU3Lpc1g6mpow=="
KEY17="OkVf3OuWJd/kg79uePp2A+ACzhc0QTIXL3zneNh91DS0QeiZyhfdm/jDIzdAnjcSm/syFz88GxvgdybS3UAwFw=="
KEY18="Ye55j6IQZb3/767RhOEpAj2xsC4DjCd5Rl1gC00mdVyYNkYP+EfdB7cn9MW+njt6E275lPZYeGmK1SNH39VAKw=="
KEY19="bbxwgWxE6tDZhkhgKuyWyT6nRt0uF0Ne/QkmCfOZCFnz0BtOXzqpgterl3/mz60IyrpFiZl0C4YFbiiJNKHPYQ=="
KEY20="//UdvxG0kX08NqFYpT4a7wqvO9cOaUeyKjNRmQNqi34g6BqsfGYnNUotgNv+9yI+U/Hgy5p0wCX5iPRoNFtT9Q=="

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

if [ ! -f ./benchmark/benchmark ]
then
    cd ./benchmark
    go build
    cd ../
    echo "Build benchmark success!"
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
    opts="--listen 127.0.0.1:$PORT --discoverpeers --datadir data$i --sealerprvkey $KEY --generate"
    if [ $i != 1 ]
    then
        opts="--norpc --listen 127.0.0.1:$PORT --discoverpeers --datadir data$i --sealerprvkey $KEY --generate"
    fi
    # send command to node window
    tmux send-keys -t cash-prototype:$i.0 "cd $SRC && ./cash-prototype $opts" ENTER
    echo "Start node with port $PORT, key $KEY and options $opts success"
done

