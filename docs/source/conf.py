# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html


# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information
project = "YukkiMusic"
project_copyright = "2025, TheTeamVivek"
author = "Vivekkumar-IN"
release = "2.0"

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = ["sphinx_copybutton", "myst_parser"]

templates_path = ["_templates"]
exclude_patterns = []

# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output

html_logo = "_static/logo-wide.svg"
html_theme = "sphinx_book_theme"
html_baseurl = "docs url"
html_static_path = ["_static"]
html_favicon = "_static/logo.png"
pygments_style = "friendly"

html_theme_options = {
#   "home_page_in_toc": True,
    "repository_url": "https://github.com/TheTeamVivek/YukkiMusic",
    "repository_branch": "dev",
    "path_to_docs": "docs",
    "use_repository_button": True,
    "use_edit_page_button": True,
    "use_issues_button": True,
    "collapse_navbar": True,
}

# html_theme_options["announcement"] = """⚠️This documentation is currently under development. Stay tuned!"""
