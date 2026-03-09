import { test, expect } from '@playwright/test';

test.describe('Todo App', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display the app header', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Todo');
  });

  test('should create a new item', async ({ page }) => {
    const itemName = `Test item ${Date.now()}`;

    // Fill the input and submit the form
    const input = page.locator('input[placeholder="What needs to be done?"]');
    await input.fill(itemName);
    await page.locator('button[type="submit"]').click();

    // Wait for the item to appear in the list
    await expect(page.locator(`text=${itemName}`)).toBeVisible({ timeout: 5000 });
  });

  test('should show Add button disabled when input is empty', async ({ page }) => {
    const addButton = page.locator('button[type="submit"]');
    await expect(addButton).toBeDisabled();
  });

  test('should enable Add button when input has text', async ({ page }) => {
    const input = page.locator('input[placeholder="What needs to be done?"]');
    await input.fill('Something');
    const addButton = page.locator('button[type="submit"]');
    await expect(addButton).toBeEnabled();
  });

  test('should clear input after adding an item', async ({ page }) => {
    const input = page.locator('input[placeholder="What needs to be done?"]');
    const itemName = `Clear test ${Date.now()}`;
    await input.fill(itemName);
    await page.locator('button[type="submit"]').click();

    // Wait for item to appear, then verify input is cleared
    await expect(page.locator(`text=${itemName}`)).toBeVisible({ timeout: 5000 });
    await expect(input).toHaveValue('');
  });

  test('should mark item as done', async ({ page }) => {
    const itemName = `Done test ${Date.now()}`;

    // Create item
    await page.locator('input[placeholder="What needs to be done?"]').fill(itemName);
    await page.locator('button[type="submit"]').click();
    await expect(page.locator(`text=${itemName}`)).toBeVisible({ timeout: 5000 });

    // Find the list item containing our text and click the toggle button (round button)
    const listItem = page.locator('li').filter({ hasText: itemName });
    await listItem.locator('button').first().click();

    // After toggling, the item should have line-through styling
    await expect(listItem.locator('span.line-through')).toBeVisible({ timeout: 5000 });
  });

  test('should filter items by active/done/all', async ({ page }) => {
    const unique = Date.now().toString();
    const activeName = `Active ${unique}`;
    const doneName = `Done ${unique}`;

    // Create two items
    await page.locator('input[placeholder="What needs to be done?"]').fill(activeName);
    await page.locator('button[type="submit"]').click();
    await expect(page.locator(`text=${activeName}`)).toBeVisible({ timeout: 5000 });

    await page.locator('input[placeholder="What needs to be done?"]').fill(doneName);
    await page.locator('button[type="submit"]').click();
    await expect(page.locator(`text=${doneName}`)).toBeVisible({ timeout: 5000 });

    // Mark the second item as done
    const doneItem = page.locator('li').filter({ hasText: doneName });
    await doneItem.locator('button').first().click();
    await expect(doneItem.locator('span.line-through')).toBeVisible({ timeout: 5000 });

    // Filter: Done — should show only the done item
    await page.locator('button', { hasText: 'Done' }).click();
    await expect(page.locator(`li >> text=${doneName}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`li >> text=${activeName}`)).not.toBeVisible();

    // Filter: Active — should show only the active item
    await page.locator('button', { hasText: 'Active' }).click();
    await expect(page.locator(`li >> text=${activeName}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`li >> text=${doneName}`)).not.toBeVisible();

    // Filter: All — should show both
    await page.locator('button', { hasText: 'All' }).click();
    await expect(page.locator(`li >> text=${activeName}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`li >> text=${doneName}`)).toBeVisible({ timeout: 5000 });
  });

  test('should search items by name', async ({ page }) => {
    const unique = Date.now().toString();

    // Create items
    await page.locator('input[placeholder="What needs to be done?"]').fill(`Buy milk ${unique}`);
    await page.locator('button[type="submit"]').click();
    await expect(page.locator(`text=Buy milk ${unique}`)).toBeVisible({ timeout: 5000 });

    await page.locator('input[placeholder="What needs to be done?"]').fill(`Walk dog ${unique}`);
    await page.locator('button[type="submit"]').click();
    await expect(page.locator(`text=Walk dog ${unique}`)).toBeVisible({ timeout: 5000 });

    // Search for "Buy" - use the search input (placeholder "Search...")
    const searchInput = page.locator('input[placeholder="Search..."]');
    await searchInput.fill('Buy');

    // Wait for filtered results
    await expect(page.locator(`li >> text=Buy milk ${unique}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`li >> text=Walk dog ${unique}`)).not.toBeVisible({ timeout: 5000 });

    // Clear search
    await searchInput.fill('');

    // All items should be visible again
    await expect(page.locator(`li >> text=Buy milk ${unique}`)).toBeVisible({ timeout: 5000 });
    await expect(page.locator(`li >> text=Walk dog ${unique}`)).toBeVisible({ timeout: 5000 });
  });

  test('should show item count in footer', async ({ page }) => {
    const itemName = `Count test ${Date.now()}`;
    await page.locator('input[placeholder="What needs to be done?"]').fill(itemName);
    await page.locator('button[type="submit"]').click();
    await expect(page.locator(`text=${itemName}`)).toBeVisible({ timeout: 5000 });

    // Footer should show "N item(s) left"
    await expect(page.locator('text=/\\d+ items? left/')).toBeVisible();
  });
});
