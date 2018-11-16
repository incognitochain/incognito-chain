cmake_minimum_required(VERSION 3.5)
project("spscoin")

set (CMAKE_CXX_STANDARD 11)

if (${CMAKE_HOST_SYSTEM_NAME} MATCHES "Darwin")
    add_definitions(-DCUSTOM_ENDIAN)
endif()

# Set build type
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

set(CMAKE_CXX_FLAGS "-Wall -Wextra")
set(CMAKE_CXX_FLAGS_DEBUG "-g")
set(CMAKE_CXX_FLAGS_RELEASE "-O3")

# Add custom CMAKE_MODULE_PATH
set(CMAKE_MODULE_PATH ${CMAKE_MODULE_PATH} "${PROJECT_SOURCE_DIR}/cmake/")

include_directories("./")
include_directories("./libsnark")

find_package(Boost REQUIRED)
find_package(sodium REQUIRED)
find_package(GMP REQUIRED)
find_package(Protobuf REQUIRED)
find_package(GRPC REQUIRED)

include_directories(${GRPC_INCLUDE_DIR})

set(EXTRA_LIBS ${EXTRA_LIBS} ${Boost_LIBRARIES})
set(EXTRA_LIBS ${EXTRA_LIBS} ${sodium_LIBRARY_RELEASE})
set(EXTRA_LIBS ${EXTRA_LIBS} ${GMP_LIBRARIES})
set(EXTRA_LIBS ${EXTRA_LIBS} ${PROTOBUF_LIBRARIES})
set(EXTRA_LIBS ${EXTRA_LIBS} ${GRPC_LIBRARY})
set(EXTRA_LIBS ${EXTRA_LIBS} ${GRPC_GRPC++_LIBRARY})

add_definitions(-DCURVE_ALT_BN128)
add_definitions(-DNO_PROCPS)
add_definitions(-DBINARY_OUTPUT)
add_definitions(-DNO_PT_COMPRESSION)
# add_definitions(-DDEBUG)
# add_definitions(-DMULTICORE)
add_definitions("-DWITH_PROCPS=OFF")

set(LS_PATH "./libsnark")
set(CURVE_PATH "${LS_PATH}/algebra/curves/alt_bn128/")

add_executable(main
               main.cpp
               JoinSplit.cpp
               Proof.cpp
               IncrementalMerkleTree.cpp
               Note.cpp
               NoteEncryption.cpp
               uint256.cpp
               util.cpp
               Address.cpp
               ${CURVE_PATH}/alt_bn128_init.cpp
               ${CURVE_PATH}/alt_bn128_pp.cpp
               ${CURVE_PATH}/alt_bn128_g1.cpp
               ${CURVE_PATH}/alt_bn128_g2.cpp
               ${CURVE_PATH}/alt_bn128_pairing.cpp
               ${LS_PATH}/common/profiling.cpp
               ${LS_PATH}/common/utils.cpp
               ../proto/zksnark.grpc.pb.cc
               ../proto/zksnark.pb.cc)
target_link_libraries(main ${EXTRA_LIBS})
