// Copyright 2021 Contributors to the Parsec project.
// SPDX-License-Identifier: Apache-2.0

package requests

// OpCode type for parsec operations
type OpCode uint32

// Operation Codes
const (
	OpPing                 OpCode = 0x0001
	OpPsaGenerateKey       OpCode = 0x0002
	OpPsaDestroyKey        OpCode = 0x0003
	OpPsaSignHash          OpCode = 0x0004
	OpPsaVerifyHash        OpCode = 0x0005
	OpPsaImportKey         OpCode = 0x0006
	OpPsaExportPublicKey   OpCode = 0x0007
	OpListProviders        OpCode = 0x0008
	OpListOpcodes          OpCode = 0x0009
	OpPsaAsymmetricEncrypt OpCode = 0x000A
	OpPsaAsymmetricDecrypt OpCode = 0x000B
	OpPsaExportKey         OpCode = 0x000C
	OpPsaGenerateRandom    OpCode = 0x000D
	OpListAuthenticators   OpCode = 0x000E
	OpPsaHashCompute       OpCode = 0x000F
	OpPsaHashCompare       OpCode = 0x0010
	OpPsaAeadEncrypt       OpCode = 0x0011
	OpPsaAeadDecrypt       OpCode = 0x0012
	OpPsaRawKeyAgreement   OpCode = 0x0013
	OpPsaCipherEncrypt     OpCode = 0x0014
	OpPsaCipherDecrypt     OpCode = 0x0015
	OpPsaMacCompute        OpCode = 0x0016
	OpPsaMacVerify         OpCode = 0x0017
	OpPsaSignMessage       OpCode = 0x0018
	OpPsaVerifyMessage     OpCode = 0x0019
	OpListKeys             OpCode = 0x001A
	OpListClients          OpCode = 0x001B
	OpDeleteClient         OpCode = 0x001C
)

func (o OpCode) IsValid() bool {
	return o <= OpDeleteClient
}
