package throttle

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/loomnetwork/go-loom"
	"github.com/loomnetwork/go-loom/auth"
	"github.com/loomnetwork/go-loom/types"
	"github.com/loomnetwork/loomchain"
	loomAuth "github.com/loomnetwork/loomchain/auth"
	"github.com/loomnetwork/loomchain/eth/utils"
	"github.com/loomnetwork/loomchain/vm"
	"github.com/pkg/errors"
)

var (
	contract = loom.MustParseAddress("chain:0x9a1aC42a17AAD6Dbc6d21c162989d0f701074044")
)

func throttleMiddlewareHandler(ttm loomchain.TxMiddlewareFunc, state loomchain.State, tx auth.SignedTx, ctx context.Context) (loomchain.TxHandlerResult, error) {
	return ttm.ProcessTx(
		state.WithContext(ctx),
		tx.Inner,
		func(state loomchain.State, txBytes []byte, isCheckTx bool) (res loomchain.TxHandlerResult, err error) {
			var nonceTx loomAuth.NonceTx
			if err := proto.Unmarshal(txBytes, &nonceTx); err != nil {
				return res, errors.Wrap(err, "throttle: unwrap nonce Tx")
			}

			var tx loomchain.Transaction
			if err := proto.Unmarshal(nonceTx.Inner, &tx); err != nil {
				return res, errors.New("throttle: unmarshal tx")
			}
			var msg vm.MessageTx
			if err := proto.Unmarshal(tx.Data, &msg); err != nil {
				return res, errors.Wrapf(err, "unmarshal message tx %v", tx.Data)
			}

			var info string
			var data []byte
			switch types.TxID(tx.Id) {
			case types.TxID_CALL:
				var callTx vm.CallTx
				if err := proto.Unmarshal(msg.Data, &callTx); err != nil {
					return res, errors.Wrapf(err, "unmarshal call tx %v", msg.Data)
				}
				if callTx.VmType == vm.VMType_EVM {
					info = utils.CallEVM
				} else {
					info = utils.CallPlugin
				}

			case types.TxID_DEPLOY:
				var deployTx vm.DeployTx
				if err := proto.Unmarshal(msg.Data, &deployTx); err != nil {
					return res, errors.Wrapf(err, "unmarshal call tx %v", msg.Data)
				}
				if deployTx.VmType == vm.VMType_EVM {
					info = utils.DeployEvm
				} else {
					info = utils.DeployPlugin
				}
				data, err = proto.Marshal(&vm.DeployResponse{
					// Always use same contract address,
					// Might want to change that later.
					Contract: contract.MarshalPB(),
				})

			case types.TxID_ETHEREUM:
				isDeploy, err := isEthDeploy(msg.Data)
				if err != nil {
					return res, err
				}
				if isDeploy {
					info = utils.DeployEvm
					data, err = proto.Marshal(&vm.DeployResponse{
						// Always use same contract address,
						// Might want to change that later.
						Contract: contract.MarshalPB(),
					})
				} else {
					info = utils.CallEVM
				}
			}
			return loomchain.TxHandlerResult{Data: data, Info: info}, err
		},
		false,
	)
}
