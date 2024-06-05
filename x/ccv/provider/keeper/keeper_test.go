package keeper_test

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"cosmossdk.io/math"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	abci "github.com/cometbft/cometbft/abci/types"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

	cryptotestutil "github.com/cosmos/interchain-security/v5/testutil/crypto"
	testkeeper "github.com/cosmos/interchain-security/v5/testutil/keeper"
	"github.com/cosmos/interchain-security/v5/x/ccv/provider/types"
	ccv "github.com/cosmos/interchain-security/v5/x/ccv/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const consumer = "consumer"

// TestValsetUpdateBlockHeight tests the getter, setter, and deletion methods for valset updates mapped to block height
func TestValsetUpdateBlockHeight(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	blockHeight, found := providerKeeper.GetValsetUpdateBlockHeight(ctx, uint64(0))
	require.False(t, found)
	require.Zero(t, blockHeight)

	providerKeeper.SetValsetUpdateBlockHeight(ctx, uint64(1), uint64(2))
	blockHeight, found = providerKeeper.GetValsetUpdateBlockHeight(ctx, uint64(1))
	require.True(t, found)
	require.Equal(t, blockHeight, uint64(2))

	providerKeeper.DeleteValsetUpdateBlockHeight(ctx, uint64(1))
	blockHeight, found = providerKeeper.GetValsetUpdateBlockHeight(ctx, uint64(1))
	require.False(t, found)
	require.Zero(t, blockHeight)

	providerKeeper.SetValsetUpdateBlockHeight(ctx, uint64(1), uint64(2))
	providerKeeper.SetValsetUpdateBlockHeight(ctx, uint64(3), uint64(4))
	blockHeight, found = providerKeeper.GetValsetUpdateBlockHeight(ctx, uint64(3))
	require.True(t, found)
	require.Equal(t, blockHeight, uint64(4))
}

// TestGetAllValsetUpdateBlockHeights tests GetAllValsetUpdateBlockHeights behaviour correctness
func TestGetAllValsetUpdateBlockHeights(t *testing.T) {
	pk, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	cases := []types.ValsetUpdateIdToHeight{
		{
			ValsetUpdateId: 2,
			Height:         22,
		},
		{
			ValsetUpdateId: 1,
			Height:         11,
		},
		{
			// normal execution should not have two ValsetUpdateIdToHeight
			// with the same Height, but let's test it anyway
			ValsetUpdateId: 4,
			Height:         11,
		},
		{
			ValsetUpdateId: 3,
			Height:         33,
		},
	}
	expectedGetAllOrder := cases
	// sorting by ValsetUpdateId
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].ValsetUpdateId < expectedGetAllOrder[j].ValsetUpdateId
	})

	for _, c := range cases {
		pk.SetValsetUpdateBlockHeight(ctx, c.ValsetUpdateId, c.Height)
	}

	// iterate and check all results are returned in the expected order
	result := pk.GetAllValsetUpdateBlockHeights(ctx)
	require.Len(t, result, len(cases))
	require.Equal(t, expectedGetAllOrder, result)
}

// TestSlashAcks tests the getter, setter, iteration, and deletion methods for stored slash acknowledgements
func TestSlashAcks(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	chainID := consumer

	acks := providerKeeper.GetSlashAcks(ctx, chainID)
	require.Nil(t, acks)

	p := []string{"alice", "bob", "charlie"}
	providerKeeper.SetSlashAcks(ctx, chainID, p)

	acks = providerKeeper.GetSlashAcks(ctx, chainID)
	require.NotNil(t, acks)

	require.Len(t, acks, 3)
	slashAcks := providerKeeper.ConsumeSlashAcks(ctx, chainID)
	require.Len(t, slashAcks, 3)

	acks = providerKeeper.GetSlashAcks(ctx, chainID)
	require.Nil(t, acks)

	chains := []string{"c1", "c2", "c3"}

	for _, c := range chains {
		providerKeeper.SetSlashAcks(ctx, c, p)
	}

	for _, c := range chains {
		require.Equal(t, p, providerKeeper.GetSlashAcks(ctx, c))
		providerKeeper.DeleteSlashAcks(ctx, c)
		acks = providerKeeper.GetSlashAcks(ctx, c)
		require.Len(t, acks, 0)
	}
}

// TestAppendSlashAck tests the append method for stored slash acknowledgements
func TestAppendSlashAck(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	p := []string{"alice", "bob", "charlie"}
	chains := []string{"c1", "c2"}
	providerKeeper.SetSlashAcks(ctx, chains[0], p)

	providerKeeper.AppendSlashAck(ctx, chains[0], p[0])
	acks := providerKeeper.GetSlashAcks(ctx, chains[0])
	require.NotNil(t, acks)
	require.Len(t, acks, len(p)+1)

	providerKeeper.AppendSlashAck(ctx, chains[1], p[0])
	acks = providerKeeper.GetSlashAcks(ctx, chains[1])
	require.NotNil(t, acks)
	require.Len(t, acks, 1)
}

// TestPendingVSCs tests the getter, appending, and deletion methods for stored pending VSCs
func TestPendingVSCs(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	chainID := consumer

	pending := providerKeeper.GetPendingVSCPackets(ctx, chainID)
	require.Len(t, pending, 0)

	_, pks, _ := ibctesting.GenerateKeys(t, 4)
	var ppks [4]tmprotocrypto.PublicKey
	for i, pk := range pks {
		ppks[i], _ = cryptocodec.ToCmtProtoPublicKey(pk)
	}

	packetList := []ccv.ValidatorSetChangePacketData{
		{
			ValidatorUpdates: []abci.ValidatorUpdate{
				{PubKey: ppks[0], Power: 1},
				{PubKey: ppks[1], Power: 2},
			},
			ValsetUpdateId: 1,
		},
		{
			ValidatorUpdates: []abci.ValidatorUpdate{
				{PubKey: ppks[2], Power: 3},
			},
			ValsetUpdateId: 2,
		},
	}
	providerKeeper.AppendPendingVSCPackets(ctx, chainID, packetList...)

	packets := providerKeeper.GetPendingVSCPackets(ctx, chainID)
	require.Len(t, packets, 2)

	newPacket := ccv.ValidatorSetChangePacketData{
		ValidatorUpdates: []abci.ValidatorUpdate{
			{PubKey: ppks[3], Power: 4},
		},
		ValsetUpdateId: 3,
	}
	providerKeeper.AppendPendingVSCPackets(ctx, chainID, newPacket)
	vscs := providerKeeper.GetPendingVSCPackets(ctx, chainID)
	require.Len(t, vscs, 3)
	require.True(t, vscs[len(vscs)-1].ValsetUpdateId == 3)
	require.True(t, vscs[len(vscs)-1].GetValidatorUpdates()[0].PubKey.String() == ppks[3].String())

	providerKeeper.DeletePendingVSCPackets(ctx, chainID)
	pending = providerKeeper.GetPendingVSCPackets(ctx, chainID)
	require.Len(t, pending, 0)
}

// TestInitHeight tests the getter and setter methods for the stored block heights (on provider) when a given consumer chain was started
func TestInitHeight(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	tc := []struct {
		chainID  string
		expected uint64
	}{
		{expected: 0, chainID: "chain"},
		{expected: 10, chainID: "chain1"},
		{expected: 12, chainID: "chain2"},
	}

	providerKeeper.SetInitChainHeight(ctx, tc[1].chainID, tc[1].expected)
	providerKeeper.SetInitChainHeight(ctx, tc[2].chainID, tc[2].expected)

	for _, tc := range tc {
		height, _ := providerKeeper.GetInitChainHeight(ctx, tc.chainID)
		require.Equal(t, tc.expected, height)
	}
}

// TestGetAllUnbondingOpIndexes tests GetAllUnbondingOpIndexes behavior correctness
func TestGetAllUnbondingOpIndexes(t *testing.T) {
	pk, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	ops := []types.VscUnbondingOps{
		{
			VscId:          2,
			UnbondingOpIds: []uint64{4, 5, 6, 7},
		},
		{
			VscId:          1,
			UnbondingOpIds: []uint64{1, 2, 3},
		},
		{
			VscId:          4,
			UnbondingOpIds: []uint64{10},
		},
		{
			VscId:          3,
			UnbondingOpIds: []uint64{8, 9},
		},
	}
	// sorting by CrossChainValidator.Address
	expectedGetAllOrder := ops
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].VscId < expectedGetAllOrder[j].VscId
	})

	pk.SetUnbondingOpIndex(ctx, "chain-2", 1, []uint64{1, 2, 3})
	for _, op := range ops {
		pk.SetUnbondingOpIndex(ctx, "chain-1", op.VscId, op.UnbondingOpIds)
	}

	// iterate and check all results are returned in the expected order
	result := pk.GetAllUnbondingOpIndexes(ctx, "chain-1")
	require.Len(t, result, len(ops))
	require.Equal(t, result, expectedGetAllOrder)
}

func TestMaturedUnbondingOps(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	ids := providerKeeper.GetMaturedUnbondingOps(ctx)
	require.Nil(t, ids)

	unbondingOpIds := []uint64{0, 1, 2, 3, 4, 5, 6}
	providerKeeper.AppendMaturedUnbondingOps(ctx, unbondingOpIds)

	ids = providerKeeper.ConsumeMaturedUnbondingOps(ctx)
	require.Equal(t, len(unbondingOpIds), len(ids))
	for i := 0; i < len(unbondingOpIds); i++ {
		require.Equal(t, unbondingOpIds[i], ids[i])
	}
}

func TestInitTimeoutTimestamp(t *testing.T) {
	pk, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	now := time.Now().UTC()
	nsNow := uint64(now.UnixNano())
	timeoutTimestamps := []types.InitTimeoutTimestamp{
		{
			ChainId:   "chain-2",
			Timestamp: nsNow,
		},
		{
			ChainId:   "chain-1",
			Timestamp: nsNow + 10,
		},
		{
			ChainId:   "chain-4",
			Timestamp: nsNow - 10,
		},
		{
			ChainId:   "chain-3",
			Timestamp: nsNow,
		},
	}

	expectedGetAllOrder := timeoutTimestamps
	// sorting by ChainId
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].ChainId < expectedGetAllOrder[j].ChainId
	})

	_, found := pk.GetInitTimeoutTimestamp(ctx, timeoutTimestamps[0].ChainId)
	require.False(t, found)

	for _, tt := range timeoutTimestamps {
		pk.SetInitTimeoutTimestamp(ctx, tt.ChainId, tt.Timestamp)
	}

	for _, tt := range timeoutTimestamps {
		_, found := pk.GetInitTimeoutTimestamp(ctx, tt.ChainId)
		require.True(t, found)
	}

	// iterate and check all results are returned in the expected order
	result := pk.GetAllInitTimeoutTimestamps(ctx)
	require.Len(t, result, len(timeoutTimestamps))
	require.Equal(t, result, expectedGetAllOrder)

	pk.DeleteInitTimeoutTimestamp(ctx, timeoutTimestamps[0].ChainId)
	_, found = pk.GetInitTimeoutTimestamp(ctx, timeoutTimestamps[0].ChainId)
	require.False(t, found)
}

// TestVscSendTimestamp tests the set, deletion, and iteration methods for VSC timeout timestamps
func TestVscSendTimestamp(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	now := time.Now().UTC()

	testCases := []struct {
		chainID string
		ts      time.Time
		vscID   uint64
	}{
		{chainID: "chain", ts: now.Add(2 * time.Hour), vscID: 2},
		{chainID: "chain", ts: now.Add(time.Hour), vscID: 1},
		{chainID: "chain", ts: now.Add(time.Hour), vscID: 3},
		// this is not possible since the ts is the timestamp of sending,
		// which means it must be in the same order as vscIDs,
		// but it still worth testing
		{chainID: "chain", ts: now.Add(-time.Hour), vscID: 4},
		{chainID: "chain1", ts: now.Add(time.Hour), vscID: 1},
		{chainID: "chain2", ts: now.Add(time.Hour), vscID: 1},
	}
	chainID := testCases[0].chainID
	expectedGetAllOrder := []types.VscSendTimestamp{}
	for _, tc := range testCases {
		if tc.chainID == chainID {
			expectedGetAllOrder = append(expectedGetAllOrder, types.VscSendTimestamp{VscId: tc.vscID, Timestamp: tc.ts})
		}
	}
	// sorting by vscID
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].VscId < expectedGetAllOrder[j].VscId
	})

	require.Empty(t, providerKeeper.GetAllVscSendTimestamps(ctx, chainID))

	for _, tc := range testCases {
		providerKeeper.SetVscSendTimestamp(ctx, tc.chainID, tc.vscID, tc.ts)
	}

	// iterate and check all results are returned in the expected order
	vscSendTimestamps := providerKeeper.GetAllVscSendTimestamps(ctx, chainID)
	require.Equal(t, expectedGetAllOrder, vscSendTimestamps)

	vscSendTimestamp, found := providerKeeper.GetFirstVscSendTimestamp(ctx, chainID)
	require.True(t, found)
	require.Equal(t, vscSendTimestamp, expectedGetAllOrder[0])

	// delete first VSC send timestamp
	providerKeeper.DeleteVscSendTimestamp(ctx, chainID, vscSendTimestamp.VscId)
	for _, vst := range providerKeeper.GetAllVscSendTimestamps(ctx, chainID) {
		require.NotEqual(t, vscSendTimestamp, vst)
	}

	// delete all VSC send timestamps
	providerKeeper.DeleteVscSendTimestampsForConsumer(ctx, chainID)
	require.Empty(t, providerKeeper.GetAllVscSendTimestamps(ctx, chainID))
}

// TestGetAllConsumerChains tests GetAllConsumerChains behaviour correctness
func TestGetAllConsumerChains(t *testing.T) {
	pk, ctx, ctrl, mocks := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	chainIDs := []string{"chain-2", "chain-1", "chain-4", "chain-3"}

	// mock the validator set
	vals := []stakingtypes.Validator{
		{OperatorAddress: "cosmosvaloper1c4k24jzduc365kywrsvf5ujz4ya6mwympnc4en"}, // 50 power
		{OperatorAddress: "cosmosvaloper196ax4vc0lwpxndu9dyhvca7jhxp70rmcvrj90c"}, // 150 power
		{OperatorAddress: "cosmosvaloper1clpqr4nrk4khgkxj78fcwwh6dl3uw4epsluffn"}, // 300 power
		{OperatorAddress: "cosmosvaloper1tflk30mq5vgqjdly92kkhhq3raev2hnz6eete3"}, // 500 power
	}
	powers := []int64{50, 150, 300, 500} // sum = 1000
	mocks.MockStakingKeeper.EXPECT().GetLastValidators(gomock.Any()).Return(vals, nil).AnyTimes()

	for i, val := range vals {
		valAddr, err := sdk.ValAddressFromBech32(val.GetOperator())
		require.NoError(t, err)
		mocks.MockStakingKeeper.EXPECT().GetLastValidatorPower(gomock.Any(), valAddr).Return(powers[i], nil).AnyTimes()
	}

	// set Top N parameters, client ids and expected result
	topNs := []uint32{0, 70, 90, 100}
	expectedMinPowerInTopNs := []int64{
		-1,  // Top N is 0, so not a Top N chain
		300, // 500 and 300 are in Top 70%
		150, // 150 is also in the top 90%,
		50,  // everyone is in the top 100%
	}

	validatorSetCaps := []uint32{0, 5, 10, 20}
	validatorPowerCaps := []uint32{0, 5, 10, 33}
	allowlists := [][]types.ProviderConsAddress{
		{},
		{types.NewProviderConsAddress([]byte("providerAddr1")), types.NewProviderConsAddress([]byte("providerAddr2"))},
		{types.NewProviderConsAddress([]byte("providerAddr3"))},
		{},
	}

	denylists := [][]types.ProviderConsAddress{
		{types.NewProviderConsAddress([]byte("providerAddr4")), types.NewProviderConsAddress([]byte("providerAddr5"))},
		{},
		{types.NewProviderConsAddress([]byte("providerAddr6"))},
		{},
	}

	expectedGetAllOrder := []types.Chain{}
	for i, chainID := range chainIDs {
		clientID := fmt.Sprintf("client-%d", len(chainIDs)-i)
		topN := topNs[i]
		pk.SetConsumerClientId(ctx, chainID, clientID)
		pk.SetTopN(ctx, chainID, topN)
		pk.SetValidatorSetCap(ctx, chainID, validatorSetCaps[i])
		pk.SetValidatorsPowerCap(ctx, chainID, validatorPowerCaps[i])
		for _, addr := range allowlists[i] {
			pk.SetAllowlist(ctx, chainID, addr)
		}
		for _, addr := range denylists[i] {
			pk.SetDenylist(ctx, chainID, addr)
		}
		strAllowlist := make([]string, len(allowlists[i]))
		for j, addr := range allowlists[i] {
			strAllowlist[j] = addr.String()
		}

		strDenylist := make([]string, len(denylists[i]))
		for j, addr := range denylists[i] {
			strDenylist[j] = addr.String()
		}

		expectedGetAllOrder = append(expectedGetAllOrder,
			types.Chain{
				ChainId:            chainID,
				ClientId:           clientID,
				Top_N:              topN,
				MinPowerInTop_N:    expectedMinPowerInTopNs[i],
				ValidatorSetCap:    validatorSetCaps[i],
				ValidatorsPowerCap: validatorPowerCaps[i],
				Allowlist:          strAllowlist,
				Denylist:           strDenylist,
			})
	}
	// sorting by chainID
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].ChainId < expectedGetAllOrder[j].ChainId
	})

	result := pk.GetAllConsumerChains(ctx)
	require.Len(t, result, len(chainIDs))
	require.Equal(t, expectedGetAllOrder, result)
}

// TestGetAllChannelToChains tests GetAllChannelToChains behaviour correctness
func TestGetAllChannelToChains(t *testing.T) {
	pk, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	chainIDs := []string{"chain-2", "chain-1", "chain-4", "chain-3"}
	expectedGetAllOrder := []types.ChannelToChain{}
	for i, chainID := range chainIDs {
		channelID := fmt.Sprintf("client-%d", len(chainIDs)-i)
		pk.SetChannelToChain(ctx, channelID, chainID)
		expectedGetAllOrder = append(expectedGetAllOrder, types.ChannelToChain{ChainId: chainID, ChannelId: channelID})
	}
	// sorting by channelID
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].ChannelId < expectedGetAllOrder[j].ChannelId
	})

	result := pk.GetAllChannelToChains(ctx)
	require.Len(t, result, len(chainIDs))
	require.Equal(t, expectedGetAllOrder, result)
}

// TestGetAllUnbondingOps tests GetAllUnbondingOps behaviour correctness
func TestGetAllUnbondingOps(t *testing.T) {
	pk, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	ops := []types.UnbondingOp{
		{
			Id:                      2,
			UnbondingConsumerChains: []string{"chain-2", "chain-1"},
		},
		{
			Id:                      1,
			UnbondingConsumerChains: []string{"chain-1", "chain-2"},
		},
		{
			Id:                      4,
			UnbondingConsumerChains: []string{"chain-2"},
		},
		{
			Id:                      3,
			UnbondingConsumerChains: []string{"chain-3", "chain-1", "chain-2"},
		},
	}
	expectedGetAllOrder := ops
	// sorting by Id
	sort.Slice(expectedGetAllOrder, func(i, j int) bool {
		return expectedGetAllOrder[i].Id < expectedGetAllOrder[j].Id
	})

	for _, op := range ops {
		pk.SetUnbondingOp(ctx, op)
	}

	// iterate and check all results are returned
	result := pk.GetAllUnbondingOps(ctx)
	require.Len(t, result, len(ops))
	require.Equal(t, expectedGetAllOrder, result)
}

// TestRemoveConsumerFromUnbondingOp tests RemoveConsumerFromUnbondingOp behaviour correctness
func TestRemoveConsumerFromUnbondingOp(t *testing.T) {
	pk, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	var expectedID uint64 = 1
	expectedUnbondingOp := types.UnbondingOp{
		Id:                      expectedID,
		UnbondingConsumerChains: []string{"chain-3", "chain-1", "chain-2"},
	}

	pk.SetUnbondingOp(ctx, expectedUnbondingOp)

	canComplete := pk.RemoveConsumerFromUnbondingOp(ctx, expectedID, "chain-1")
	require.False(t, canComplete)
	unbondingOp, found := pk.GetUnbondingOp(ctx, expectedID)
	require.True(t, found)
	expectedChainIDs := []string{"chain-3", "chain-2"}
	require.Equal(t, expectedChainIDs, unbondingOp.UnbondingConsumerChains)

	canComplete = pk.RemoveConsumerFromUnbondingOp(ctx, expectedID, "chain-2")
	require.False(t, canComplete)
	unbondingOp, found = pk.GetUnbondingOp(ctx, expectedID)
	require.True(t, found)
	expectedChainIDs = []string{"chain-3"}
	require.Equal(t, expectedChainIDs, unbondingOp.UnbondingConsumerChains)

	// check that it doesn't panic when calling with same chain ID
	canComplete = pk.RemoveConsumerFromUnbondingOp(ctx, expectedID, "chain-2")
	require.False(t, canComplete)
	unbondingOp, found = pk.GetUnbondingOp(ctx, expectedID)
	require.True(t, found)
	require.Equal(t, expectedChainIDs, unbondingOp.UnbondingConsumerChains)

	canComplete = pk.RemoveConsumerFromUnbondingOp(ctx, expectedID, "chain-3")
	require.True(t, canComplete)
	unbondingOp, found = pk.GetUnbondingOp(ctx, expectedID)
	require.False(t, found)
	require.Empty(t, unbondingOp.UnbondingConsumerChains)

	// check that it panics when calling with wrong chain IDs
	require.Panics(t, func() {
		canComplete = pk.RemoveConsumerFromUnbondingOp(ctx, expectedID, "some_chain")
		require.False(t, canComplete)
	})
}

// TestSetSlashLog tests slash log getter and setter methods
func TestSetSlashLog(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	addrWithDoubleSigns := cryptotestutil.NewCryptoIdentityFromIntSeed(1).ProviderConsAddress()
	addrWithoutDoubleSigns := cryptotestutil.NewCryptoIdentityFromIntSeed(2).ProviderConsAddress()

	providerKeeper.SetSlashLog(ctx, addrWithDoubleSigns)
	require.True(t, providerKeeper.GetSlashLog(ctx, addrWithDoubleSigns))
	require.False(t, providerKeeper.GetSlashLog(ctx, addrWithoutDoubleSigns))
}

func TestSetProposedConsumerChains(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	tests := []struct {
		chainID    string
		proposalID uint64
	}{
		{chainID: "1", proposalID: 1},
		{chainID: "some other ID", proposalID: 12},
		{chainID: "some other other chain ID", proposalID: 123},
		{chainID: "", proposalID: 1234},
	}

	for _, test := range tests {
		providerKeeper.SetProposedConsumerChain(ctx, test.chainID, test.proposalID)
		cID, _ := providerKeeper.GetProposedConsumerChain(ctx, test.proposalID)
		require.Equal(t, cID, test.chainID)
	}
}

func TestDeleteProposedConsumerChainInStore(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	tests := []struct {
		chainID          string
		proposalID       uint64
		deleteProposalID uint64
		ok               bool
	}{
		{chainID: "1", proposalID: 1, deleteProposalID: 1, ok: true},
		{chainID: "", proposalID: 12, deleteProposalID: 12, ok: true},
		{chainID: "1", proposalID: 0, deleteProposalID: 1, ok: false},
	}
	for _, test := range tests {
		providerKeeper.SetProposedConsumerChain(ctx, test.chainID, test.proposalID)
		providerKeeper.DeleteProposedConsumerChainInStore(ctx, test.deleteProposalID)
		cID, found := providerKeeper.GetProposedConsumerChain(ctx, test.proposalID)
		if test.ok {
			require.False(t, found)
		} else {
			require.Equal(t, cID, test.chainID)
		}
	}
}

func TestGetAllProposedConsumerChainIDs(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()
	tests := [][]types.ProposedChain{
		{},
		{
			{
				ChainID:    "1",
				ProposalID: 1,
			},
		},
		{
			{
				ChainID:    "1",
				ProposalID: 1,
			},
			{
				ChainID:    "2",
				ProposalID: 2,
			},
			{
				ChainID:    "",
				ProposalID: 3,
			},
		},
	}

	for _, test := range tests {
		for _, tc := range test {
			providerKeeper.SetProposedConsumerChain(ctx, tc.ChainID, tc.ProposalID)
		}

		chains := providerKeeper.GetAllProposedConsumerChainIDs(ctx)

		sort.Slice(chains, func(i, j int) bool {
			return chains[i].ProposalID < chains[j].ProposalID
		})
		sort.Slice(test, func(i, j int) bool {
			return test[i].ProposalID < test[j].ProposalID
		})
		require.Equal(t, chains, test)

		for _, tc := range test {
			providerKeeper.DeleteProposedConsumerChainInStore(ctx, tc.ProposalID)
		}
	}
}

// TestTopN tests the `SetTopN`, `GetTopN`, `IsTopN`, and `IsOptIn` methods
func TestTopN(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	tests := []struct {
		chainID string
		N       uint32
		isOptIn bool
	}{
		{chainID: "TopNChain1", N: 50, isOptIn: false},
		{chainID: "TopNChain2", N: 100, isOptIn: false},
		{chainID: "OptInChain", N: 0, isOptIn: true},
	}

	for _, test := range tests {
		providerKeeper.SetTopN(ctx, test.chainID, test.N)
		topN, found := providerKeeper.GetTopN(ctx, test.chainID)
		require.Equal(t, test.N, topN)
		require.True(t, found)

		if test.isOptIn {
			require.True(t, providerKeeper.IsOptIn(ctx, test.chainID))
			require.False(t, providerKeeper.IsTopN(ctx, test.chainID))
		} else {
			require.False(t, providerKeeper.IsOptIn(ctx, test.chainID))
			require.True(t, providerKeeper.IsTopN(ctx, test.chainID))
		}
	}
}

func TestGetAllOptedIn(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	expectedOptedInValidators := []types.ProviderConsAddress{
		types.NewProviderConsAddress([]byte("providerAddr1")),
		types.NewProviderConsAddress([]byte("providerAddr2")),
		types.NewProviderConsAddress([]byte("providerAddr3")),
	}

	for _, expectedOptedInValidator := range expectedOptedInValidators {
		providerKeeper.SetOptedIn(ctx, "chainID", expectedOptedInValidator)
	}

	actualOptedInValidators := providerKeeper.GetAllOptedIn(ctx, "chainID")

	// sort validators first to be able to compare
	sortOptedInValidators := func(addresses []types.ProviderConsAddress) {
		sort.Slice(addresses, func(i, j int) bool {
			return bytes.Compare(addresses[i].ToSdkConsAddr(), addresses[j].ToSdkConsAddr()) < 0
		})
	}
	sortOptedInValidators(expectedOptedInValidators)
	sortOptedInValidators(actualOptedInValidators)
	require.Equal(t, expectedOptedInValidators, actualOptedInValidators)
}

// TestOptedIn tests the `SetOptedIn`, `IsOptedIn`, `DeleteOptedIn` and `DeleteAllOptedIn` methods
func TestOptedIn(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	optedInValidator1 := types.NewProviderConsAddress([]byte("providerAddr1"))
	optedInValidator2 := types.NewProviderConsAddress([]byte("providerAddr2"))

	require.False(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator1))
	providerKeeper.SetOptedIn(ctx, "chainID", optedInValidator1)
	require.True(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator1))
	providerKeeper.DeleteOptedIn(ctx, "chainID", optedInValidator1)
	require.False(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator1))

	providerKeeper.SetOptedIn(ctx, "chainID", optedInValidator1)
	providerKeeper.SetOptedIn(ctx, "chainID", optedInValidator2)
	require.True(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator1))
	require.True(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator2))
	providerKeeper.DeleteAllOptedIn(ctx, "chainID")
	require.False(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator1))
	require.False(t, providerKeeper.IsOptedIn(ctx, "chainID", optedInValidator2))
}

// TestConsumerCommissionRate tests the `SetConsumerCommissionRate`, `GetConsumerCommissionRate`, and `DeleteConsumerCommissionRate` methods
func TestConsumerCommissionRate(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	providerAddr1 := types.NewProviderConsAddress([]byte("providerAddr1"))
	providerAddr2 := types.NewProviderConsAddress([]byte("providerAddr2"))

	cr, found := providerKeeper.GetConsumerCommissionRate(ctx, "chainID", providerAddr1)
	require.False(t, found)
	require.Equal(t, math.LegacyZeroDec(), cr)

	providerKeeper.SetConsumerCommissionRate(ctx, "chainID", providerAddr1, math.LegacyOneDec())
	cr, found = providerKeeper.GetConsumerCommissionRate(ctx, "chainID", providerAddr1)
	require.True(t, found)
	require.Equal(t, math.LegacyOneDec(), cr)

	providerKeeper.SetConsumerCommissionRate(ctx, "chainID", providerAddr2, math.LegacyZeroDec())
	cr, found = providerKeeper.GetConsumerCommissionRate(ctx, "chainID", providerAddr2)
	require.True(t, found)
	require.Equal(t, math.LegacyZeroDec(), cr)

	provAddrs := providerKeeper.GetAllCommissionRateValidators(ctx, "chainID")
	require.Len(t, provAddrs, 2)

	for _, addr := range provAddrs {
		providerKeeper.DeleteConsumerCommissionRate(ctx, "chainID", addr)
	}

	_, found = providerKeeper.GetConsumerCommissionRate(ctx, "chainID", providerAddr1)
	require.False(t, found)

	_, found = providerKeeper.GetConsumerCommissionRate(ctx, "chainID", providerAddr2)
	require.False(t, found)
}

// TestValidatorsPowerCap tests the `SetValidatorsPowerCap`, `GetValidatorsPowerCap`, and `DeleteValidatorsPowerCap` methods
func TestValidatorsPowerCap(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	expectedPowerCap := uint32(10)
	providerKeeper.SetValidatorsPowerCap(ctx, "chainID", expectedPowerCap)
	powerCap, found := providerKeeper.GetValidatorsPowerCap(ctx, "chainID")
	require.Equal(t, expectedPowerCap, powerCap)
	require.True(t, found)

	providerKeeper.DeleteValidatorsPowerCap(ctx, "chainID")
	_, found = providerKeeper.GetValidatorsPowerCap(ctx, "chainID")
	require.False(t, found)
}

// TestValidatorSetCap tests the `SetValidatorSetCap`, `GetValidatorSetCap`, and `DeleteValidatorSetCap` methods
func TestValidatorSetCap(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	expectedPowerCap := uint32(10)
	providerKeeper.SetValidatorSetCap(ctx, "chainID", expectedPowerCap)
	powerCap, found := providerKeeper.GetValidatorSetCap(ctx, "chainID")
	require.Equal(t, expectedPowerCap, powerCap)
	require.True(t, found)

	providerKeeper.DeleteValidatorSetCap(ctx, "chainID")
	_, found = providerKeeper.GetValidatorSetCap(ctx, "chainID")
	require.False(t, found)
}

// TestAllowlist tests the `SetAllowlist`, `IsAllowlisted`, `DeleteAllowlist`, and `IsAllowlistEmpty` methods
func TestAllowlist(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	chainID := "chainID"

	// no validator was allowlisted and hence the allowlist is empty
	require.True(t, providerKeeper.IsAllowlistEmpty(ctx, chainID))

	providerAddr1 := types.NewProviderConsAddress([]byte("providerAddr1"))
	providerKeeper.SetAllowlist(ctx, chainID, providerAddr1)
	require.True(t, providerKeeper.IsAllowlisted(ctx, chainID, providerAddr1))

	// allowlist is not empty anymore
	require.False(t, providerKeeper.IsAllowlistEmpty(ctx, chainID))

	providerAddr2 := types.NewProviderConsAddress([]byte("providerAddr2"))
	providerKeeper.SetAllowlist(ctx, chainID, providerAddr2)
	require.True(t, providerKeeper.IsAllowlisted(ctx, chainID, providerAddr2))
	require.False(t, providerKeeper.IsAllowlistEmpty(ctx, chainID))

	providerKeeper.DeleteAllowlist(ctx, chainID)
	require.False(t, providerKeeper.IsAllowlisted(ctx, chainID, providerAddr1))
	require.False(t, providerKeeper.IsAllowlisted(ctx, chainID, providerAddr2))
	require.True(t, providerKeeper.IsAllowlistEmpty(ctx, chainID))
}

// TestDenylist tests the `SetDenylist`, `IsDenylisted`, `DeleteDenylist`, and `IsDenylistEmpty` methods
func TestDenylist(t *testing.T) {
	providerKeeper, ctx, ctrl, _ := testkeeper.GetProviderKeeperAndCtx(t, testkeeper.NewInMemKeeperParams(t))
	defer ctrl.Finish()

	chainID := "chainID"

	// no validator was denylisted and hence the denylist is empty
	require.True(t, providerKeeper.IsDenylistEmpty(ctx, chainID))

	providerAddr1 := types.NewProviderConsAddress([]byte("providerAddr1"))
	providerKeeper.SetDenylist(ctx, chainID, providerAddr1)
	require.True(t, providerKeeper.IsDenylisted(ctx, chainID, providerAddr1))

	// denylist is not empty anymore
	require.False(t, providerKeeper.IsDenylistEmpty(ctx, chainID))

	providerAddr2 := types.NewProviderConsAddress([]byte("providerAddr2"))
	providerKeeper.SetDenylist(ctx, chainID, providerAddr2)
	require.True(t, providerKeeper.IsDenylisted(ctx, chainID, providerAddr2))
	require.False(t, providerKeeper.IsDenylistEmpty(ctx, chainID))

	providerKeeper.DeleteDenylist(ctx, chainID)
	require.False(t, providerKeeper.IsDenylisted(ctx, chainID, providerAddr1))
	require.False(t, providerKeeper.IsDenylisted(ctx, chainID, providerAddr2))
	require.True(t, providerKeeper.IsDenylistEmpty(ctx, chainID))
}
