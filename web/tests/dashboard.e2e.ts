import { test } from '@e2e-dev/web';
import { expect } from 'e2e';

test('dashboard header and navigation tabs render correctly', async ({ app, screen }) => {
  await app.open('/');

  // Verify header brand elements
  await expect(screen.getByText('AIM')).toBeVisible();
  await expect(screen.getByText('Agent Identity Manager')).toBeVisible();

  // Verify navigation tabs
  const profilesTab = screen.getByRole('tab', { name: /Profiles/i });
  const sessionsTab = screen.getByRole('tab', { name: /Sessions/i });
  const mcpTab = screen.getByRole('tab', { name: /MCP Servers/i });

  await expect(profilesTab).toBeVisible();
  await expect(sessionsTab).toBeVisible();
  await expect(mcpTab).toBeVisible();

  // Navigate to Sessions tab
  await sessionsTab.tap();
  await expect(screen.getByPlaceholder('Search sessions by goal, cwd, title, or profile...')).toBeVisible();

  // Navigate to MCP Servers tab
  await mcpTab.tap();
  await expect(screen.getByText('Model Context Protocol (MCP) Infrastructure')).toBeVisible();

  // Return to Profiles tab
  await profilesTab.tap();
  await expect(screen.getByRole('button', { name: /Add Profile/i })).toBeVisible();
});

test('profile filtering and add modal workflow', async ({ app, screen }) => {
  await app.open('/');

  // Verify Add Profile button exists
  const addBtn = screen.getByRole('button', { name: /Add Profile/i });
  await expect(addBtn).toBeVisible();

  // Open Add Profile dialog
  await addBtn.tap();
  await expect(screen.getByText('Create New Agent Profile')).toBeVisible();

  // Verify modal inputs exist
  const nameInput = screen.getByPlaceholder('e.g. work, sandbox, client-x');
  await expect(nameInput).toBeVisible();

  // Fill in test profile name
  await nameInput.fill('e2e-test-profile');

  // Close dialog by clicking Cancel
  const cancelBtn = screen.getByRole('button', { name: /Cancel/i });
  await cancelBtn.tap();
  await expect(screen.getByText('Create New Agent Profile')).not.toBeVisible();
});
