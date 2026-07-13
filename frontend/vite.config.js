import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite';
import Components from 'unplugin-vue-components/vite';
import { NaiveUiResolver } from 'unplugin-vue-components/resolvers';
import { TDesignResolver } from '@tdesign-vue-next/auto-import-resolver';

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
      vue(),
      AutoImport({
          resolvers: [TDesignResolver({
              library: 'chat'
          })],
      }),
      Components({
          resolvers: [
              NaiveUiResolver(),
              TDesignResolver({
                  library: 'chat'
              }),
          ],
      }),
  ]
})
