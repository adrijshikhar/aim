import { test } from '@e2e-dev/web';
import { expect } from 'e2e';

test('agent explores dashboard and switches views via natural language', async ({ app, agent, screen }) => {
  const isLiveConfigured = Boolean(process.env.TINKER_API_KEY || process.env.OPENAI_API_KEY);
  if (!isLiveConfigured) {
    test.skip('Set TINKER_API_KEY to run live Tinker API Labs agent steps');
  }

  await app.open('/');

  // Agent drives the UI using natural language goals
  await agent.act('navigate to the sessions tab');
  await agent.assert('the sessions table is visible');

  await agent.act('navigate to the mcp servers tab');
  await agent.assert('the MCP servers section is visible');

  await agent.act('return to the profiles tab');
  await expect(screen.getByRole('button', { name: /Add Profile/i })).toBeVisible();
});
