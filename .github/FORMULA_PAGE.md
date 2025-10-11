# Homebrew Formula Page

This repository includes a GitHub Actions workflow that generates a simple GitHub Pages site for OpenScribe using Jekyll's default theme.

## 🌐 Live Site

Once deployed, the formula page will be available at:
**https://alexandrelam.github.io/openscribe/**

## 🚀 How to Generate/Update the Page

### Option 1: Via GitHub UI (Easiest)

1. Go to the **Actions** tab: https://github.com/alexandrelam/openscribe/actions
2. Click on **"Deploy Formula Page"** workflow in the left sidebar
3. Click the **"Run workflow"** button (top right)
4. Optionally add a reason (e.g., "Updated to v0.2.0")
5. Click **"Run workflow"**
6. Wait ~1-2 minutes for the build to complete
7. Visit: https://alexandrelam.github.io/openscribe/

### Option 2: Via GitHub CLI

```bash
# Trigger the workflow
gh workflow run deploy-formula-page.yml -f reason="Updated formula page"

# Check the workflow status
gh run list --workflow=deploy-formula-page.yml
```

## ⚙️ First-Time Setup

### 1. Enable GitHub Pages

Go to: https://github.com/alexandrelam/openscribe/settings/pages

- **Source**: Select branch `gh-pages` and folder `/ (root)`
- Click **Save**

### 2. Enable Workflow Permissions

Go to: https://github.com/alexandrelam/openscribe/settings/actions

- Under "Workflow permissions", select **"Read and write permissions"**
- Click **Save**

### 3. Run the Workflow

Follow the steps in "How to Generate/Update the Page" above.

## 📁 Files

- `.github/workflows/deploy-formula-page.yml` - GitHub Actions workflow
- `docs/index.md` - Main content (markdown)
- `docs/_config.yml` - Jekyll configuration
- `.github/FORMULA_PAGE.md` - This file

## 🎨 Customization

### Change Content

Edit `docs/index.md` to update the page content. It's just markdown, so it's easy to edit!

### Change Theme

Edit `docs/_config.yml` and change the `theme` value. Available themes:
- `minima` (default, clean and minimal)
- `minimal`
- `cayman`
- `slate`
- `modernist`

See all themes: https://pages.github.com/themes/

## 🔧 How It Works

When you run the workflow:

1. Checks out this repository
2. Copies the `docs/` folder to the `gh-pages` branch
3. GitHub Pages automatically uses Jekyll to render the markdown as HTML

## 📋 When to Regenerate

You should manually run the workflow when:

- ✅ You update `docs/index.md` with new content
- ✅ You release a new version
- ✅ You change installation instructions
- ✅ You want to update the page styling/theme

**Note:** The workflow does NOT run automatically on commits. You must trigger it manually.

## 🐛 Troubleshooting

### "Page not found" error

- Verify GitHub Pages is enabled (Settings → Pages)
- Check that the `gh-pages` branch exists
- Wait a few minutes for GitHub's CDN to update
- Clear your browser cache

### Workflow fails

Check the Actions tab for errors. Common issues:

1. **Permissions error**
   - Solution: Enable "Read and write permissions" in Settings → Actions

2. **Page looks unstyled**
   - Solution: Make sure `docs/_config.yml` exists with a theme specified
   - GitHub Pages needs a few minutes to build the Jekyll site

## 💡 Tips

- The workflow is lightweight and completes in ~1-2 minutes
- You can run it as many times as you want
- It's safe to run multiple times - it just overwrites the existing page
- The `gh-pages` branch is auto-created on first run
- Edit `docs/index.md` to change the page content anytime

## 🔗 Links

- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Jekyll Themes](https://pages.github.com/themes/)

---

**Questions?** Open an issue or check the main README.md
