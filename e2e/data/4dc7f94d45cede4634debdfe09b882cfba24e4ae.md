# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: tests/site-add-to-cart.spec.ts >> Site - Add to Cart >> should add product to cart from product detail page
- Location: e2e/tests/site-add-to-cart.spec.ts:30:7

# Error details

```
Error: expect(received).toBe(expected) // Object.is equality

Expected: 1
Received: 0
```

# Page snapshot

```yaml
- generic [ref=e3]:
  - banner [ref=e4]:
    - banner [ref=e5]:
      - generic [ref=e6]:
        - link "Home" [ref=e7] [cursor=pointer]:
          - /url: /
          - img [ref=e9]
        - generic [ref=e11]:
          - generic [ref=e12]:
            - button "Switch to English" [ref=e13] [cursor=pointer]: EN
            - button "Switch to 中文" [ref=e14] [cursor=pointer]: ZH
            - button "Switch to 한국어" [ref=e15] [cursor=pointer]: KO
          - link "CART (1)" [ref=e16] [cursor=pointer]:
            - /url: /cart
            - button "CART (1)" [ref=e17]:
              - generic [ref=e18]:
                - img [ref=e19]
                - generic [ref=e21]: CART (1)
  - main [ref=e22]:
    - generic [ref=e25]:
      - heading "YOUR CART" [level=1] [ref=e27]
      - generic [ref=e29]:
        - generic [ref=e30]:
          - heading "ITEMS (1)" [level=2] [ref=e31]
          - generic [ref=e33]:
            - generic [ref=e35]:
              - heading "Test Product 1" [level=3] [ref=e36]
              - generic [ref=e38]: $10.99
            - generic [ref=e39]:
              - generic [ref=e40]:
                - button "Decrease quantity" [disabled] [ref=e41]: "-"
                - textbox "Quantity" [ref=e42]: "1"
                - button "Increase quantity" [ref=e43] [cursor=pointer]: +
              - button "REMOVE" [ref=e44] [cursor=pointer]
        - generic [ref=e46]:
          - generic [ref=e47]: TOTAL
          - generic [ref=e48]: $10.99
        - paragraph [ref=e50]: NO PAYMENT SYSTEMS AVAILABLE. CONTACT ADMINISTRATOR.
  - contentinfo [ref=e51]:
    - contentinfo [ref=e52]:
      - generic [ref=e53]:
        - navigation [ref=e55]:
          - link "Terms & Conditions" [ref=e56] [cursor=pointer]:
            - /url: /terms
          - link "Privacy Policy" [ref=e57] [cursor=pointer]:
            - /url: /privacy
          - link "Cookies" [ref=e58] [cursor=pointer]:
            - /url: /cookies
        - generic [ref=e60]:
          - paragraph [ref=e61]: © 2026 All Rights Reserved
          - link "Powered by myCart" [ref=e62] [cursor=pointer]:
            - /url: https://github.com/shurco/mycart
  - dialog "COOKIE CONSENT" [ref=e63]:
    - generic [ref=e65]:
      - generic [ref=e66]:
        - heading "COOKIE CONSENT" [level=2] [ref=e67]
        - paragraph [ref=e68]:
          - text: We use cookies to enhance your browsing experience and analyze site traffic. By clicking 'Accept', you consent to our use of cookies.
          - link "Learn more in our Privacy Policy" [ref=e69] [cursor=pointer]:
            - /url: /privacy
      - generic [ref=e70]:
        - button "REJECT" [ref=e71] [cursor=pointer]
        - button "ACCEPT" [ref=e72] [cursor=pointer]
```

# Test source

```ts
  1  | import { Page } from 'patchright'
  2  | import { expect } from 'patchright/test'
  3  | 
  4  | /**
  5  |  * Feature Object for Shopping Cart Page (Storefront)
  6  |  * Handles cart viewing and product removal
  7  |  */
  8  | export class CartFeature {
  9  |   constructor(private page: Page) {}
  10 | 
  11 |   async goto() {
  12 |     await this.page.goto('/cart')
  13 |   }
  14 | 
  15 |   async waitForCart() {
  16 |     // Wait for the cart page to fully load - wait for either items or empty message
  17 |     await this.page.waitForSelector('[data-testid="cart-item"], text=/cart is empty/i', { timeout: 10000 })
  18 |   }
  19 | 
  20 |   async verifyPageLoaded() {
  21 |     await expect(this.page).toHaveURL('/cart')
  22 |     await this.waitForCart()
  23 |   }
  24 | 
  25 |   async getCartItemCount() {
  26 |     const items = await this.page.locator('[data-testid="cart-item"]').count()
  27 |     return items
  28 |   }
  29 | 
  30 |   async verifyCartIsEmpty() {
  31 |     // Use getByRole to target only the h1 heading, not the paragraph text
  32 |     const emptyMessage = this.page.getByRole('heading', { name: /cart is empty/i })
  33 |     await expect(emptyMessage).toBeVisible()
  34 |   }
  35 | 
  36 |   async verifyCartHasItems(count: number) {
  37 |     // Wait a moment for cart to load from localStorage
  38 |     await this.page.waitForTimeout(1000)
  39 |     const items = await this.getCartItemCount()
> 40 |     expect(items).toBe(count)
     |                   ^ Error: expect(received).toBe(expected) // Object.is equality
  41 |   }
  42 | 
  43 |   async getItemNameByIndex(index: number) {
  44 |     const item = this.page.locator('[data-testid="cart-item"]').nth(index)
  45 |     const name = await item.locator('[data-testid="item-name"]').textContent()
  46 |     return name?.trim() || ''
  47 |   }
  48 | 
  49 |   async removeItemByIndex(index: number) {
  50 |     const item = this.page.locator('[data-testid="cart-item"]').nth(index)
  51 |     const removeButton = item.locator('button:has-text("REMOVE"), button[aria-label*="remove" i]')
  52 |     await removeButton.click()
  53 |   }
  54 | 
  55 |   async verifyItemRemoved(productName: string) {
  56 |     const item = this.page.locator(`[data-testid="cart-item"]:has-text("${productName}")`)
  57 |     await expect(item).not.toBeVisible()
  58 |   }
  59 | 
  60 |   async getTotalPrice() {
  61 |     const total = await this.page.locator('[data-testid="cart-total"]').textContent()
  62 |     return total?.trim() || ''
  63 |   }
  64 | }
  65 | 
```