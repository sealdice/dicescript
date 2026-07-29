// Webpack entry point - imports gopherjs compiled output
import _module from './dicescript.cjs';

// gopherjs exports to module.exports.ds, not module.exports directly
const diceModule = _module.ds || _module;

const {
  newVM,
  newVMForPlaygournd,
  help,
  vmNewDict,
  vmNewFloat,
  vmNewInt,
  vmNewStr,
  newValueMap,
  newConfig,
} = diceModule;

export {
  newVM,
  newVMForPlaygournd,
  help,
  vmNewDict,
  vmNewFloat,
  vmNewInt,
  vmNewStr,
  newValueMap,
  newConfig,
};

export const ds = diceModule;

export default diceModule;
