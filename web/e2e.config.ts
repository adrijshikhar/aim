import type { E2EConfig } from 'e2e';
import { web } from '@e2e-dev/web';
import { createOpenAICompatible } from '@ai-sdk/openai-compatible';

const tinkerApiKey = process.env.TINKER_API_KEY || process.env.OPENAI_API_KEY || 'tinker-key';
const tinkerBaseUrl = process.env.TINKER_API_BASE_URL || 'https://api.tinker.thinkingmachines.ai/v1';
const tinkerModel = process.env.TINKER_MODEL || 'gpt-4o';

const tinker = createOpenAICompatible({
  name: 'tinker',
  baseURL: tinkerBaseUrl,
  apiKey: tinkerApiKey,
});

export default {
  targets: [
    {
      engine: web(),
      app: {
        url: process.env.APP_URL || 'http://127.0.0.1:8080',
        command: {
          executable: 'go',
          args: ['run', 'github.com/aim-cli/aim/cmd/aim', 'web', '--port', '8080', '--no-open'],
          reuseExisting: true,
          log: '.e2e/logs/app.log',
        },
      },
    },
  ],
  agents: {
    default: {
      model: tinker.chatModel(tinkerModel),
      system: 'You are an expert QA testing agent verifying the AIM Web Dashboard.',
    },
  },
} satisfies E2EConfig;
