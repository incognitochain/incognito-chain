APP_NAME="constant-bootnode"

file="$GOPATH/bin/$APP_NAME"
echo $file
if [ -f "$file" ]
then
    echo "Bootnode app is built in GOPATH:'$GOPATH'"
else
    echo "Bootnode app not exist, need to be built"
    sh build.sh
fi


port=9330

while [ "$1" != "" ]; do
    case $1 in
        -p | --rpcport )           shift
                                port=$1
                                ;;
        -h | --help )           usage
                                exit
                                ;;
        * )                     usage
                                exit 1
    esac
    shift
done

echo $file
if [ -f "$file" ]
then
    echo "Running boothnode .............................."
    $GOPATH/bin/$APP_NAME -p $port
else
    echo "Build failure"
fi
