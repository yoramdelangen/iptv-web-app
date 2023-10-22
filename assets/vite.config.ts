import { resolve, join, } from 'node:path';
import { existsSync, renameSync, unlinkSync } from 'node:fs';

const outDir = '../public';
const templateDir = '../templates';

/** @type {import('vite').UserConfig} */
export default {
  build: {
    outDir,
  },
  plugins: [
    {
      name: 'move-index-file-somewhere-else',
      closeBundle: async () => {

        const absOut = resolve(outDir);
        const templateOut = resolve(templateDir);

        const index = join(absOut, 'index.html');
        const templateIndex = join(templateOut, 'layout.html');

        if (existsSync(templateIndex)) {
          unlinkSync(templateIndex);
        }

        if (existsSync(index)) {
          renameSync(index, templateIndex);
        }

        if (existsSync(templateIndex)) {
          console.log('We moved it!!', templateIndex)
        } else {
          console.log('Not moved index.html')
        }
      }
    },
  ]
}
