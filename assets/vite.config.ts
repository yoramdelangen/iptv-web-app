import { exec } from 'node:child_process'
import { resolve, join, } from 'node:path';
import { existsSync, unlinkSync, rmSync } from 'node:fs';

const outDir = './dist';
const assetDir = '../public/assets';
const templateDir = '../layouts';

/** @type {import('vite').UserConfig} */
export default {
  build: {
    // outDir,
  },
  plugins: [
    {
      name: 'move-index-file-somewhere-else',
      closeBundle: async () => {
        const absOut = resolve(outDir);
        const assetsOut = resolve(assetDir);
        const templateOut = resolve(templateDir);

        const index = join(absOut, 'index.html');
        const templateIndex = join(templateOut, 'main.html');

        if (existsSync(templateIndex)) {
          unlinkSync(templateIndex);
        }

        if (existsSync(assetsOut)) {
          console.log('Removed assets folder')
          rmSync(assetsOut, { recursive: true, force: true });
        }

        if (existsSync(index)) {
          console.log('Moving index.html file')
          exec(['cp', index, templateIndex].join(' '))
        }

        exec(['cp', '-R', absOut + '/assets', assetsOut].join(' '))
        console.log('Moved assets')
      }
    },
  ]
}
