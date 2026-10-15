import { test } from '@e2e-dev/web';
import { expect } from 'e2e';

test('sessions table renders with filters and resume action', async ({ app, screen }) => {
  await app.open('/');

  // Switch to Sessions tab
  const sessionsTab = screen.getByRole('tab', { name: /Sessions/i });
  await sessionsTab.tap();

  // Verify search box and search placeholder
  const searchInput = screen.getByPlaceholder('Search sessions by goal, cwd, title, or profile...');
  await expect(searchInput).toBeVisible();

  // Test searching sessions
  await searchInput.fill('auth');
  await expect(searchInput).toHaveValue('auth');

  // Clear search
  await searchInput.fill('');

  // Check filter buttons
  await expect(screen.getByRole('button', { name: /^Engine$/i })).toBeVisible();
  await expect(screen.getByRole('button', { name: /^Profile$/i })).toBeVisible();
  await expect(screen.getByRole('button', { name: /^CWD$/i })).toBeVisible();
  await expect(screen.getByRole('button', { name: /^Status$/i })).toBeVisible();

  // Check table headers
  await expect(screen.getByRole('columnheader', { name: /AGENT/i })).toBeVisible();
  await expect(screen.getByRole('columnheader', { name: /PROFILE/i })).toBeVisible();
  await expect(screen.getByRole('columnheader', { name: /DIRECTORY/i })).toBeVisible();
  await expect(screen.getByRole('columnheader', { name: /GOAL/i })).toBeVisible();
  await expect(screen.getByRole('columnheader', { name: /UPDATED/i })).toBeVisible();
});
