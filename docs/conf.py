# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html


# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information
project = "YukkiMusic"
project_copyright = "2024-2025, TheTeamVivek"
author = "TheTeamVivek"
release = "2.0"

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

extensions = ["sphinx_copybutton", "myst_parser"]

# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output

html_title = "YukkiMusic 2.0"
html_logo = "_static/logo-wide.svg"
html_theme = "sphinx_book_theme"
html_baseurl = "https://TheTeamVivek.github.io"
html_static_path = ["_static"]
html_css_files = ["custom.css"]
html_favicon = "_static/logo.png"
pygments_style = "friendly"

html_theme_options = {
    "announcement": "⚠️This documentation is currently under development. Stay tuned!",
    "use_repository_button": True,
    "use_edit_page_button": True,
    "use_issues_button": True,
    "use_source_button": True,
    "use_fullscreen_button": False,
    "home_page_in_toc": True,
    "path_to_docs": "docs",
    "repository_url": "https://github.com/TheTeamVivek/YukkiMusic",
    "repository_branch": "dev",
    "footer_content_items": "copyright.html, last-updated.html, extra-footer.html",
    "search_bar_text": "Search the docs",
    "back_to_top_button": False,
}