import diceModule from './dist/main.mjs';

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
