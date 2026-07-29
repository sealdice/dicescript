import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

export default {
  entry: './entry.js',
  mode: 'production',
  experiments: {
    outputModule: true,
  },
  output: {
    filename: 'main.mjs',
    path: path.resolve(__dirname, 'dist'),
    library: {
      type: 'module',
    },
  },
};
