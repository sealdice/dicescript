// Webpack entry point - imports gopherjs compiled output
import diceModule from './dicescript.cjs';

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
