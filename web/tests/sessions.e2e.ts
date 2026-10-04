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

  // Check table headers
  await expect(screen.getByText('Agent')).toBeVisible();
  await expect(screen.getByText('Profile')).toBeVisible();
  await expect(screen.getByText('Directory (CWD)')).toBeVisible();
  await expect(screen.getByText('Goal / Title')).toBeVisible();
  await expect(screen.getByText('Action')).toBeVisible();
});
