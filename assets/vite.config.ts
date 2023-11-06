import { exec } from 'node:child_process'
import { resolve, join, } from 'node:path';
import { existsSync, unlinkSync, rmSync } from 'node:fs';

const outDir = './dist';
const assetDir = '../internal/statics/assets';
const templateDir = '../templates/layouts';
const isProduction = process.env.NODE_ENV === 'production';

const absOut = resolve(outDir);
const assetsOut = resolve(assetDir);
const templateOut = resolve(templateDir);


console.log('Running production:', isProduction)

/** @type {import('vite').UserConfig} */
export default {
  build: {
    // outDir,
    rollupOptions: {
      output: {
        assetFileNames() {
          if (isProduction) {
            return 'assets/[name]-[hash][extname]'
          }
          return 'assets/[name][extname]';
        },
        chunkFileNames() {
          if (isProduction) {
            return '[name]-[hash].js'
          }
          return '[name].js'
        }
      }
    }
  },
  plugins: [
    {
      name: 'move-index-file-somewhere-else',
      closeBundle: async () => {
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
