echo "Start build bootnode"

echo "go get"
go get

sleep 7 &
PID=$!
echo $PID
i=1
sp="/-\|"
echo -n ' '
while [ -d /proc/$PID ]
do
  printf "\b${sp:i++%${#sp}:1}"
done

echo "go build -o bootnode"
go build -o bootnode

sleep 7 &
PID=$!
i=1
sp="/-\|"
echo -n ' '
while [ -d /proc/$PID ]
do
  printf "\b${sp:i++%${#sp}:1}"
done

echo "Build bootnode success!"
