<script lang="ts">
  import { goto } from '$app/navigation'
  import { base } from '$app/paths'
  import { onDestroy, onMount } from 'svelte'
  import { browser } from '$app/environment'
  import Blank from '$lib/layouts/Blank.svelte'
  import FormInput from '$lib/components/form/Input.svelte'
  import FormButton from '$lib/components/form/Button.svelte'
  import { apiPost, apiGet } from '$lib/utils/api'
  import { showMessage } from '$lib/utils/message'
  import { translate } from '$lib/i18n'

  // Reactive translation function
  let t = $derived($translate)

  let email = $state('')
  let password = $state('')
  let domain = $state('')
  let emailError = $state('')
  let passwordError = $state('')
  let domainError = $state('')

  // Cloudflare-specific fields
  let isCloudflare = $state(false)
  let cfAccountID = $state('')
  let cfAPIToken = $state('')
  let cfD1DatabaseID = $state('')
  let cfR2BucketName = $state('')
  let d1Databases = $state<Array<{uuid: string, name: string}>>([])
  let r2Buckets = $state<Array<{name: string}>>([])
  let loadingD1 = $state(false)
  let loadingR2 = $state(false)
  let showCreateD1 = $state(false)
  let showCreateR2 = $state(false)
  let newD1Name = $state('')
  let newR2Name = $state('')

  let redirectTimer: ReturnType<typeof setTimeout> | undefined
  onDestroy(() => {
    if (redirectTimer !== undefined) clearTimeout(redirectTimer)
  })

  function validateEmail(value: string) {
    if (!value) {
      return t('install.emailRequired')
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
      return t('install.emailInvalid')
    }
    return ''
  }

  function validatePassword(value: string) {
    if (!value) {
      return t('install.passwordRequired')
    }
    if (value.length < 6) {
      return t('install.passwordMinLength')
    }
    if (value.length > 72) {
      return t('install.passwordMaxLength')
    }
    return ''
  }

  function validateDomain(value: string) {
    if (!value) {
      return t('install.domainRequired')
    }
    // Basic domain validation
    const domainRegex = /^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$/i
    if (!domainRegex.test(value) && value !== 'localhost' && !value.startsWith('localhost:')) {
      return t('install.domainInvalid')
    }
    return ''
  }

  async function handleSubmit(event?: Event) {
    event?.preventDefault()

    emailError = validateEmail(email)
    passwordError = validatePassword(password)
    domainError = validateDomain(domain)

    if (emailError || passwordError || domainError) {
      return
    }

    try {
      const payload: any = { email, password, domain }

      // Include Cloudflare config if in Cloudflare mode
      if (isCloudflare) {
        payload.cf_account_id = cfAccountID
        payload.cf_api_token = cfAPIToken
        payload.cf_d1_database_id = cfD1DatabaseID
        payload.cf_r2_bucket_name = cfR2BucketName
      }

      const res = await apiPost(`/api/install`, payload)
      if (res?.success) {
        showMessage(t('install.installedSuccessfully'), 'connextSuccess')
        // Redirect to signin page after successful installation
        redirectTimer = setTimeout(() => {
          goto(`${base}/signin`)
        }, 1000)
      } else {
        showMessage(res?.result || res?.message || t('install.installationFailed'), 'connextError')
      }
    } catch (error) {
      showMessage(t('install.networkError'), 'connextError')
    }
  }

  async function loadD1Databases() {
    loadingD1 = true
    try {
      const res = await apiGet(`/api/install/cloudflare/d1/databases?account_id=${encodeURIComponent(cfAccountID)}&api_token=${encodeURIComponent(cfAPIToken)}`)
      d1Databases = res?.result || []
    } catch (error) {
      console.error('Failed to load D1 databases:', error)
    } finally {
      loadingD1 = false
    }
  }

  async function loadR2Buckets() {
    loadingR2 = true
    try {
      const res = await apiGet(`/api/install/cloudflare/r2/buckets?account_id=${encodeURIComponent(cfAccountID)}&api_token=${encodeURIComponent(cfAPIToken)}`)
      r2Buckets = res?.result || []
    } catch (error) {
      console.error('Failed to load R2 buckets:', error)
    } finally {
      loadingR2 = false
    }
  }

  async function createD1Database() {
    if (!newD1Name.trim()) return
    try {
      const res = await apiPost('/api/install/cloudflare/d1/databases', {
        name: newD1Name,
        account_id: cfAccountID,
        api_token: cfAPIToken
      })
      if (res?.error) {
        showMessage(res.error, 'connextError')
      } else if (res?.result?.uuid) {
        cfD1DatabaseID = res.result.uuid
        await loadD1Databases()
        showCreateD1 = false
        newD1Name = ''
        showMessage('D1 database created', 'connextSuccess')
      } else {
        showMessage('Failed to create D1 database', 'connextError')
      }
    } catch (error) {
      showMessage('Failed to create D1 database', 'connextError')
    }
  }

  async function createR2Bucket() {
    if (!newR2Name.trim()) return
    try {
      const res = await apiPost('/api/install/cloudflare/r2/buckets', {
        name: newR2Name,
        account_id: cfAccountID,
        api_token: cfAPIToken
      })
      if (res?.error) {
        showMessage(res.error, 'connextError')
      } else if (res?.result?.name) {
        cfR2BucketName = res.result.name
        await loadR2Buckets()
        showCreateR2 = false
        newR2Name = ''
        showMessage('R2 bucket created', 'connextSuccess')
      } else {
        showMessage('Failed to create R2 bucket', 'connextError')
      }
    } catch (error) {
      showMessage('Failed to create R2 bucket', 'connextError')
    }
  }

  async function loadCloudflareCredentials() {
    if (cfAccountID && cfAPIToken) {
      await Promise.all([loadD1Databases(), loadR2Buckets()])
    }
  }

  onMount(async () => {
    // Set default domain from current location if in browser
    if (browser) {
      const url = new URL(window.location.href)
      domain = url.origin.replace(/^https?:\/\//, '')

      // Cloudflare mode detection removed - maintenance API no longer exists
      // User can manually enable Cloudflare mode if needed
      isCloudflare = false
    }
  })
</script>

<Blank>
  <div class="mx-auto max-w-screen-xl px-4 py-16 sm:px-6 lg:px-8">
    <div class="mx-auto max-w-lg text-center">
      <h1 class="text-2xl font-bold sm:text-3xl">🛒 {t('install.title')} myCart</h1>
      <p class="mt-4 text-gray-600">{t('install.configureCart')}</p>
    </div>
    <form onsubmit={(e) => handleSubmit(e)} class="mx-auto mt-8 mb-0 max-w-md space-y-4">
      <FormInput id="email" type="email" title={t('install.email')} ico="at-symbol" error={emailError} bind:value={email} />
      <FormInput
        id="password"
        type="password"
        title={t('install.password')}
        ico="finger-print"
        error={passwordError}
        bind:value={password}
      />
      <FormInput
        id="domain"
        type="text"
        title={t('install.domain')}
        ico="glob-alt"
        error={domainError}
        bind:value={domain}
        placeholder="example.com"
      />

      {#if isCloudflare}
        <div class="rounded-lg border-2 border-blue-200 bg-blue-50 p-4 space-y-3">
          <div class="flex items-center gap-2">
            <span class="text-lg">☁️</span>
            <h3 class="font-semibold text-blue-900">Cloudflare Configuration</h3>
          </div>

          <FormInput
            id="cf_account_id"
            type="text"
            title="Cloudflare Account ID"
            placeholder="your-account-id"
            bind:value={cfAccountID}
          />
          <FormInput
            id="cf_api_token"
            type="password"
            title="Cloudflare API Token"
            placeholder="your-api-token"
            bind:value={cfAPIToken}
          />

          {#if cfAccountID && cfAPIToken}
            <button
              type="button"
              onclick={() => loadCloudflareCredentials()}
              class="text-sm text-blue-600 hover:text-blue-800 underline"
            >
              Load D1 Databases and R2 Buckets
            </button>

            <!-- D1 Database Selection -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">D1 Database</label>
              {#if !showCreateD1}
                <div class="flex gap-2">
                  <select
                    bind:value={cfD1DatabaseID}
                    class="flex-1 rounded-md border-gray-300 shadow-sm"
                    disabled={loadingD1}
                  >
                    <option value="">Select a D1 database...</option>
                    {#each d1Databases as db}
                      <option value={db.uuid}>{db.name}</option>
                    {/each}
                  </select>
                  <button
                    type="button"
                    onclick={() => (showCreateD1 = true)}
                    class="px-3 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
                  >
                    Create New
                  </button>
                </div>
              {:else}
                <div class="flex gap-2">
                  <input
                    type="text"
                    bind:value={newD1Name}
                    placeholder="my-database"
                    class="flex-1 rounded-md border-gray-300 shadow-sm"
                  />
                  <button
                    type="button"
                    onclick={createD1Database}
                    class="px-3 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
                  >
                    Create
                  </button>
                  <button
                    type="button"
                    onclick={() => (showCreateD1 = false)}
                    class="px-3 py-2 text-sm bg-gray-400 text-white rounded-md hover:bg-gray-500"
                  >
                    Cancel
                  </button>
                </div>
              {/if}
            </div>

            <!-- R2 Bucket Selection -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">R2 Bucket</label>
              {#if !showCreateR2}
                <div class="flex gap-2">
                  <select
                    bind:value={cfR2BucketName}
                    class="flex-1 rounded-md border-gray-300 shadow-sm"
                    disabled={loadingR2}
                  >
                    <option value="">Select an R2 bucket...</option>
                    {#each r2Buckets as bucket}
                      <option value={bucket.name}>{bucket.name}</option>
                    {/each}
                  </select>
                  <button
                    type="button"
                    onclick={() => (showCreateR2 = true)}
                    class="px-3 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
                  >
                    Create New
                  </button>
                </div>
              {:else}
                <div class="flex gap-2">
                  <input
                    type="text"
                    bind:value={newR2Name}
                    placeholder="my-bucket"
                    class="flex-1 rounded-md border-gray-300 shadow-sm"
                  />
                  <button
                    type="button"
                    onclick={createR2Bucket}
                    class="px-3 py-2 text-sm bg-green-600 text-white rounded-md hover:bg-green-700"
                  >
                    Create
                  </button>
                  <button
                    type="button"
                    onclick={() => (showCreateR2 = false)}
                    class="px-3 py-2 text-sm bg-gray-400 text-white rounded-md hover:bg-gray-500"
                  >
                    Cancel
                  </button>
                </div>
              {/if}
            </div>
          {:else}
            <p class="text-sm text-blue-700">
              Enter Account ID and API Token above, then click "Load" to see available databases and buckets.
            </p>
          {/if}

          <p class="text-sm text-blue-700">
            💡 D1 database will be initialized automatically with migrations during installation.
          </p>
        </div>
      {/if}

      <FormButton type="submit" name={t('install.installButton')} color="green" ico="arrow-right" />
    </form>
  </div>
</Blank>
