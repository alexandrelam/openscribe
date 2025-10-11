# Homebrew Formula Page

This repository includes a GitHub Actions workflow that generates a beautiful Homebrew formula page for OpenScribe.

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
6. Wait ~2-3 minutes for the build to complete
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
- `.github/formula-page-template.mustache` - HTML template
- `.github/FORMULA_PAGE.md` - This file

## 🎨 Customization

### Change Colors

Edit `.github/formula-page-template.mustache` and modify the CSS variables:

```css
:root {
  --primary-color: #FF6B6B;      /* Change to your brand color */
  --secondary-color: #4ECDC4;    /* Change accent color */
  /* ... */
}
```

### Change Content

The page automatically pulls information from your Homebrew formula:
- Name, description, version
- Dependencies
- Bottles (pre-built binaries)

To add custom sections, edit the template file.

## 🔧 How It Works

When you run the workflow:

1. Checks out this repository
2. Installs your Homebrew formula from the tap
3. Extracts formula information as JSON
4. Uses the Mustache template to generate HTML
5. Pushes the generated HTML to the `gh-pages` branch
6. GitHub Pages serves the static site

## 📋 When to Regenerate

You should manually run the workflow when:

- ✅ You release a new version
- ✅ You update the formula description
- ✅ You change dependencies
- ✅ You want to update the page styling
- ✅ You add new bottles/platforms

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

2. **Formula not found**
   - Solution: Ensure your formula exists at `alexandrelam/openscribe/openscribe`
   - Test locally: `brew tap alexandrelam/openscribe && brew info alexandrelam/openscribe/openscribe`

3. **JSON parsing error**
   - Solution: Verify your formula is valid: `brew audit alexandrelam/openscribe/openscribe`

## 💡 Tips

- The workflow is lightweight and completes in ~2-3 minutes
- You can run it as many times as you want
- It's safe to run multiple times - it just overwrites the existing page
- The `gh-pages` branch is auto-created on first run

## 🔗 Links

- [GitHub Pages Documentation](https://docs.github.com/en/pages)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)

---

**Questions?** Open an issue or check the main README.md
