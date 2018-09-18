#include "Note.hpp"
// #include "prf.h"
// #include "crypto/sha256.h"

// #include "random.h"
// #include "version.h"
// #include "streams.h"

// #include "util.h"
// #include "librustzcash.h"

using namespace libzcash;

SproutNote::SproutNote() {
    a_pk = random_uint256();
    rho = random_uint256();
    r = random_uint256();
}

// // Construct and populate Sapling note for a given payment address and value.
// SaplingNote::SaplingNote(const SaplingPaymentAddress& address, const uint64_t value) : BaseNote(value) {
//     d = address.d;
//     pk_d = address.pk_d;
//     librustzcash_sapling_generate_r(r.begin());
// }

// // Call librustzcash to compute the commitment
// boost::optional<uint256> SaplingNote::cm() const {
//     uint256 result;
//     if (!librustzcash_sapling_compute_cm(
//             d.data(),
//             pk_d.begin(),
//             value(),
//             r.begin(),
//             result.begin()
//         ))
//     {
//         return boost::none;
//     }

//     return result;
// }

// // Call librustzcash to compute the nullifier
// boost::optional<uint256> SaplingNote::nullifier(const SaplingSpendingKey& sk, const uint64_t position) const
// {
//     auto vk = sk.full_viewing_key();
//     auto ak = vk.ak;
//     auto nk = vk.nk;

//     uint256 result;
//     if (!librustzcash_sapling_compute_nf(
//             d.data(),
//             pk_d.begin(),
//             value(),
//             r.begin(),
//             ak.begin(),
//             nk.begin(),
//             position,
//             result.begin()
//     ))
//     {
//         return boost::none;
//     }

//     return result;
// }

SproutNotePlaintext::SproutNotePlaintext(
    const SproutNote& note,
    std::array<unsigned char, ZC_MEMO_SIZE> memo) : BaseNotePlaintext(note, memo)
{
    rho = note.rho;
    r = note.r;
}

SproutNote SproutNotePlaintext::note(const SproutPaymentAddress& addr) const
{
    return SproutNote(addr.a_pk, value_, rho, r);
}

SproutNotePlaintext SproutNotePlaintext::decrypt(const ZCNoteDecryption& decryptor,
                                     const ZCNoteDecryption::Ciphertext& ciphertext,
                                     const uint256& ephemeralKey,
                                     const uint256& h_sig,
                                     unsigned char nonce
                                    )
{
    // auto plaintext = decryptor.decrypt(ciphertext, ephemeralKey, h_sig, nonce);

    // CDataStream ss(SER_NETWORK, PROTOCOL_VERSION);
    // ss << plaintext;

    SproutNotePlaintext ret;
    // ss >> ret;

    // assert(ss.size() == 0);

    return ret;
}

ZCNoteEncryption::Ciphertext SproutNotePlaintext::encrypt(ZCNoteEncryption& encryptor,
                                                    const uint256& pk_enc
                                                   ) const
{
    // CDataStream ss(SER_NETWORK, PROTOCOL_VERSION);
    // ss << (*this);

    ZCNoteEncryption::Plaintext pt;

    // assert(pt.size() == ss.size());

    // memcpy(&pt[0], &ss[0], pt.size());

    return encryptor.encrypt(pk_enc, pt);
}
