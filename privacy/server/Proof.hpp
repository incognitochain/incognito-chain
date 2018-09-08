#ifndef ZC_PROOF_H_
#define ZC_PROOF_H_

// #include "serialize.h"
#include "uint256.h"

const unsigned char G1_PREFIX_MASK = 0x02;
const unsigned char G2_PREFIX_MASK = 0x0a;

namespace libzcash {

// Element in the base field
class Fq {
private:

public:
    base_blob<256> data;
    Fq() : data() { }

    template<typename libsnark_Fq>
    Fq(libsnark_Fq element);

    template<typename libsnark_Fq>
    libsnark_Fq to_libsnark_fq() const;

    std::string to_string() {
        return std::string(data.begin(), data.end());
    }

    bool from_string(std::string &s) {
        if (s.size() != 32)  // 32 bytes
            return false;
        const unsigned char *data_mem = (const unsigned char *)s.c_str();
        memcpy(data.begin(), data_mem, s.size());
        return true;
    }

    // ADD_SERIALIZE_METHODS;

    // template <typename Stream, typename Operation>
    // inline void SerializationOp(Stream& s, Operation ser_action) {
    //     READWRITE(data);
    // }

    friend bool operator==(const Fq& a, const Fq& b)
    {
        return (
            a.data == b.data
        );
    }

    friend bool operator!=(const Fq& a, const Fq& b)
    {
        return !(a == b);
    }
};

// Element in the extension field
class Fq2 {
private:

public:
    base_blob<512> data;
    Fq2() : data() { }

    template<typename libsnark_Fq2>
    Fq2(libsnark_Fq2 element);

    template<typename libsnark_Fq2>
    libsnark_Fq2 to_libsnark_fq2() const;

    std::string to_string() {
        return std::string(data.begin(), data.end());
    }

    bool from_string(std::string &s) {
        if (s.size() != 64)  // 64 bytes
            return false;
        const unsigned char *data_mem = (const unsigned char *)s.c_str();
        memcpy(data.begin(), data_mem, s.size());
        return true;
    }

    // ADD_SERIALIZE_METHODS;

    // template <typename Stream, typename Operation>
    // inline void SerializationOp(Stream& s, Operation ser_action) {
    //     READWRITE(data);
    // }

    friend bool operator==(const Fq2& a, const Fq2& b)
    {
        return (
            a.data == b.data
        );
    }

    friend bool operator!=(const Fq2& a, const Fq2& b)
    {
        return !(a == b);
    }
};

// Compressed point in G1
class CompressedG1 {
private:
    bool y_lsb;
    Fq x;

public:
    CompressedG1() : y_lsb(false), x() { }

    template<typename libsnark_G1>
    CompressedG1(libsnark_G1 point);

    template<typename libsnark_G1>
    libsnark_G1 to_libsnark_g1() const;

    void print() {
        char leadingByte[2] = {0};
        leadingByte[0] |= y_lsb;
        printf("%d ", leadingByte[0]);
        for (int i = 0; i < 32; ++i)
            printf("%d ", x.data.begin()[i]);
        printf("\n");
    }

    std::string to_string() {
        unsigned char leadingByte = 0;
        leadingByte |= y_lsb;
        return std::string(1, leadingByte) + x.to_string();
    }

    bool from_string(const std::string &s) {
        if (s.size() != 33) { // 32 bytes and y_lsb
            return false;
        }
        const unsigned char *data_mem = (const unsigned char *)s.c_str();
        unsigned char leadingByte = data_mem[0];
        if (leadingByte > 1) {
            return false;
        }

        y_lsb = leadingByte & 1;
        std::string sub = s.substr(1);
        return x.from_string(sub);
    }

    // ADD_SERIALIZE_METHODS;

    // template <typename Stream, typename Operation>
    // inline void SerializationOp(Stream& s, Operation ser_action) {
    //     unsigned char leadingByte = G1_PREFIX_MASK;

    //     if (y_lsb) {
    //         leadingByte |= 1;
    //     }

    //     READWRITE(leadingByte);

    //     if ((leadingByte & (~1)) != G1_PREFIX_MASK) {
    //         throw std::ios_base::failure("lead byte of G1 point not recognized");
    //     }

    //     y_lsb = leadingByte & 1;

    //     READWRITE(x);
    // }

    friend bool operator==(const CompressedG1& a, const CompressedG1& b)
    {
        return (
            a.y_lsb == b.y_lsb &&
            a.x == b.x
        );
    }

    friend bool operator!=(const CompressedG1& a, const CompressedG1& b)
    {
        return !(a == b);
    }
};

// Compressed point in G2
class CompressedG2 {
private:
    bool y_gt;
    Fq2 x;

public:
    CompressedG2() : y_gt(false), x() { }

    template<typename libsnark_G2>
    CompressedG2(libsnark_G2 point);

    template<typename libsnark_G2>
    libsnark_G2 to_libsnark_g2() const;

    void print() {
        unsigned char leadingByte[2] = {0};
        leadingByte[0] |= y_gt;
        printf("%d ", leadingByte[0]);
        for (int i = 0; i < 64; ++i)
            printf("%d ", x.data.begin()[i]);
        printf("\n");
    }

    std::string to_string() {
        unsigned char leadingByte = 0;
        leadingByte |= y_gt;
        std::string result = std::string(1, leadingByte) + x.to_string();
        return result;
    }

    bool from_string(const std::string &s) {
        if (s.size() != 65) { // 64 bytes and y_gt
            return false;
        }
        const unsigned char *data_mem = (const unsigned char *)s.c_str();
        unsigned char leadingByte = data_mem[0];
        if (leadingByte > 1) {
            return false;
        }

        y_gt = leadingByte & 1;
        std::string sub = s.substr(1);
        return x.from_string(sub);
    }

    // ADD_SERIALIZE_METHODS;

    // template <typename Stream, typename Operation>
    // inline void SerializationOp(Stream& s, Operation ser_action) {
    //     unsigned char leadingByte = G2_PREFIX_MASK;

    //     if (y_gt) {
    //         leadingByte |= 1;
    //     }

    //     READWRITE(leadingByte);

    //     if ((leadingByte & (~1)) != G2_PREFIX_MASK) {
    //         throw std::ios_base::failure("lead byte of G2 point not recognized");
    //     }

    //     y_gt = leadingByte & 1;

    //     READWRITE(x);
    // }

    friend bool operator==(const CompressedG2& a, const CompressedG2& b)
    {
        return (
            a.y_gt == b.y_gt &&
            a.x == b.x
        );
    }

    friend bool operator!=(const CompressedG2& a, const CompressedG2& b)
    {
        return !(a == b);
    }
};

// Compressed zkSNARK proof
class PHGRProof {
public:
    CompressedG1 g_A;
    CompressedG1 g_A_prime;
    CompressedG2 g_B;
    CompressedG1 g_B_prime;
    CompressedG1 g_C;
    CompressedG1 g_C_prime;
    CompressedG1 g_K;
    CompressedG1 g_H;

    PHGRProof() : g_A(), g_A_prime(), g_B(), g_B_prime(), g_C(), g_C_prime(), g_K(), g_H() { }

    // Produces a compressed proof using a libsnark zkSNARK proof
    template<typename libsnark_proof>
    PHGRProof(const libsnark_proof& proof);

    // Produces a libsnark zkSNARK proof out of this proof,
    // or throws an exception if it is invalid.
    template<typename libsnark_proof>
    libsnark_proof to_libsnark_proof() const;

    static PHGRProof random_invalid();

    // ADD_SERIALIZE_METHODS;

    // template <typename Stream, typename Operation>
    // inline void SerializationOp(Stream& s, Operation ser_action) {
    //     READWRITE(g_A);
    //     READWRITE(g_A_prime);
    //     READWRITE(g_B);
    //     READWRITE(g_B_prime);
    //     READWRITE(g_C);
    //     READWRITE(g_C_prime);
    //     READWRITE(g_K);
    //     READWRITE(g_H);
    // }

    friend bool operator==(const PHGRProof& a, const PHGRProof& b)
    {
        return (
            a.g_A == b.g_A &&
            a.g_A_prime == b.g_A_prime &&
            a.g_B == b.g_B &&
            a.g_B_prime == b.g_B_prime &&
            a.g_C == b.g_C &&
            a.g_C_prime == b.g_C_prime &&
            a.g_K == b.g_K &&
            a.g_H == b.g_H
        );
    }

    friend bool operator!=(const PHGRProof& a, const PHGRProof& b)
    {
        return !(a == b);
    }
};

void initialize_curve_params();
}
#endif // ZC_PROOF_H_
