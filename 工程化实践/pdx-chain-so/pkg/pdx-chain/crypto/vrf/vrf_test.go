/*************************************************************************
 * Copyright (C) 2016-2019 PDX Technologies, Inc. All Rights Reserved.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 * @Time   : 2020/10/15 10:08 上午
 * @Author : liangc
 *************************************************************************/

package vrf

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"math/big"
	"pdx-chain/crypto"
	"testing"
)

// DEMO
func TestVRFForS256(t *testing.T) {
	var (
		prvkeyHex = "98f6d85b289687be58b59741a096d7a0feae1c5ae396dbca6ec61814684b5d27"
		prvkey, _ = func() (*ecdsa.PrivateKey, error) {
			prvkeyBuf, _ := hex.DecodeString(prvkeyHex)
			return crypto.ToECDSA(prvkeyBuf)
		}()
		pubkey = &prvkey.PublicKey
		msg    = []byte("hello world.")
		// 生成 VRF 随机数
		genRandom = func(k *ecdsa.PrivateKey, msg []byte) (n [32]byte, proof []byte) {
			if vrfPrvkey, err := NewVRFSigner(k); err != nil {
				panic(err)
			} else {
				return vrfPrvkey.Evaluate(msg)
			}
		}
		// 验证 VRF 随机数
		verifyRandom = func(pub *ecdsa.PublicKey, msg []byte, n [32]byte, proof []byte) error {
			if vrfPubkey, err := NewVRFVerifier(pub); err != nil {
				return err
			} else if _n, err := vrfPubkey.ProofToHash(msg, proof); err != nil {
				return err
			} else if _n == n {
				return nil
			}
			return errors.New("vrf verify fail")
		}

		print = func(msg []byte, n [32]byte, p []byte, r error) {
			t.Log("---------------------------------------------->")
			t.Log("msg:", string(msg))
			t.Log("random:", new(big.Int).SetBytes(n[:]))
			t.Log("proof:", crypto.Keccak256Hash(p).Hex())
			t.Log("result:", r == nil)
			t.Log("----------------------------------------------<\n")
		}
	)

	// 使用相同的私钥和原文，反复生成随机数，将会得到相同的 n 和不同的 p
	for i := 0; i < 3; i++ {
		// 用私钥 prvkey 加原文 msg 生成随机数 n 和证据 p
		n, p := genRandom(prvkey, msg)
		t.Log("proof len:", len(p))
		// 用公钥 pubkey 对原文 msg 随机数 n 和证据 p 进行验证
		r := verifyRandom(pubkey, msg, n, p)
		print(msg, n, p, r)
	}
}
