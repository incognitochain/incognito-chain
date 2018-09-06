#ifndef ZC_NOTE_H_
#define ZC_NOTE_H_

#include "uint256.h"
#include "Zcash.h"
#include "Address.hpp"
#include "NoteEncryption.hpp"

#include <array>
#include <boost/optional.hpp>

namespace libzcash {

class BaseNote {
public:
    BaseNote() {}
    BaseNote(uint64_t value) : value_(value) {};
    virtual ~BaseNote() {};

    uint64_t value_ = 0;
    inline uint64_t value() const { return value_; };
};

class SproutNote : public BaseNote {
public:
    uint256 a_pk;
    uint256 rho;
    uint256 r;
    uint256 cm;
    uint256 nf;

    SproutNote(uint256 a_pk, uint64_t value, uint256 rho, uint256 r)
        : BaseNote(value), a_pk(a_pk), rho(rho), r(r) {}

    SproutNote();

    virtual ~SproutNote() {};

    // uint256 cm() const;

    // uint256 nullifier(const SproutSpendingKey& a_sk) const;
};


// class SaplingNote : public BaseNote {
// public:
//     diversifier_t d;
//     uint256 pk_d;
//     uint256 r;

//     SaplingNote(diversifier_t d, uint256 pk_d, uint64_t value, uint256 r)
//             : BaseNote(value), d(d), pk_d(pk_d), r(r) {}

//     SaplingNote() {};

//     SaplingNote(const SaplingPaymentAddress &address, uint64_t value);

//     virtual ~SaplingNote() {};

//     boost::optional<uint256> cm() const;
//     boost::optional<uint256> nullifier(const SaplingSpendingKey &sk, const uint64_t position) const;
// };

class BaseNotePlaintext {
protected:
    uint64_t value_ = 0;
    std::array<unsigned char, ZC_MEMO_SIZE> memo_;
public:
    BaseNotePlaintext() {}
    BaseNotePlaintext(const BaseNote& note, std::array<unsigned char, ZC_MEMO_SIZE> memo)
        : value_(note.value()), memo_(memo) {}
    virtual ~BaseNotePlaintext() {}

    inline uint64_t value() const { return value_; }
    inline const std::array<unsigned char, ZC_MEMO_SIZE> & memo() const { return memo_; }
};

class SproutNotePlaintext : public BaseNotePlaintext {
public:
    uint256 rho;
    uint256 r;

    SproutNotePlaintext() {}

    SproutNotePlaintext(const SproutNote& note, std::array<unsigned char, ZC_MEMO_SIZE> memo);

    SproutNote note(const SproutPaymentAddress& addr) const;

    virtual ~SproutNotePlaintext() {}

    // ADD_SERIALIZE_METHODS;

    // template <typename Stream, typename Operation>
    // inline void SerializationOp(Stream& s, Operation ser_action) {
    //     unsigned char leadingByte = 0x00;
    //     READWRITE(leadingByte);

    //     if (leadingByte != 0x00) {
    //         throw std::ios_base::failure("lead byte of SproutNotePlaintext is not recognized");
    //     }

    //     READWRITE(value_);
    //     READWRITE(rho);
    //     READWRITE(r);
    //     READWRITE(memo_);
    // }

    static SproutNotePlaintext decrypt(const ZCNoteDecryption& decryptor,
                                 const ZCNoteDecryption::Ciphertext& ciphertext,
                                 const uint256& ephemeralKey,
                                 const uint256& h_sig,
                                 unsigned char nonce
                                );

    ZCNoteEncryption::Ciphertext encrypt(ZCNoteEncryption& encryptor,
                                         const uint256& pk_enc
                                        ) const;
};

}

#endif // ZC_NOTE_H_
