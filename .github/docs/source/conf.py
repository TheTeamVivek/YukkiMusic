# Configuration file for the Sphinx documentation builder.
#
# For the full list of built-in configuration values, see the documentation:
# https://www.sphinx-doc.org/en/master/usage/configuration.html

# -- Project information -----------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#project-information

import os
import sys


sys.path.insert(0, os.path.abspath("../.."))

project = "YukkiMusic"
copyright = "2024-2025, TheTeamVivek"
author = "TheTeamVivek"

# -- General configuration ---------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#general-configuration

autosummary_generate = True

extensions = [
    "sphinx.ext.autosummary",
    "sphinx.ext.autodoc",
    "sphinx.ext.napoleon",
    "sphinx_copybutton",
    "sphinx.ext.intersphinx",
    "sphinx_reredirects",
]
templates_path = ["_templates"]
exclude_patterns = []


# -- Options for HTML output -------------------------------------------------
# https://www.sphinx-doc.org/en/master/usage/configuration.html#options-for-html-output
html_title = project
html_theme = "sphinx_rtd_theme"

html_theme_options = {
    "navigation_with_keys": True,
}
html_copy_source = False
html_static_path = ["_static"]
html_extra_path = ["_templates"]
html_favicon = html_static_path[0] + "/TheTeamVivek.ico"


napoleon_include_special_with_doc = False
napoleon_numpy_docstring = False
napoleon_include_special_with_doc = False
napoleon_use_rtype = False
napoleon_use_param = True

autodoc_member_order = "groupwise"

intersphinx_mapping = {"python": ("https://docs.python.org/3", None)}

redirects = {
    "chk": "https://docs.python.org/3",
}

html_context = {
    "display_github": True,
    "github_user": "TheTeamVivek",
    "github_repo": "YukkiMusic",
    "github_version": "dev",
    "conf_py_path": "/.github/docs/source/",
}
