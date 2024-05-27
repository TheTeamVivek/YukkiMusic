"""import os
import shutil
import git


def copy_plugins(source, destination):
    # Iterate over all files and folders in the source directory
    for item in os.listdir(source):
        source_path = os.path.join(source, item)
        destination_path = os.path.join(destination, item)
        # If it's a file, copy it to the destination folder
        if os.path.isfile(source_path):
            shutil.copy2(source_path, destination_path)
        # If it's a folder, recursively copy its contents
        elif os.path.isdir(source_path):
            shutil.copytree(source_path, destination_path)


def load_external_plugin(
    repo_url="https://github.com/Vivekkumar-IN/External-Plugins",
    main_repo_path="YukkiMusic/plugins/tools",
):
    # Clone the repository to a temporary folder
    temp_repo_path = "cache"
    git.Repo.clone_from(repo_url, temp_repo_path)
    # Navigate to the plugins folder in the cloned repository
    source_plugins_path = os.path.join(temp_repo_path, "plugins")
    # Copy the contents of the plugins folder to the main repository
    copy_plugins(source_plugins_path, main_repo_path)
    # Remove the temporary cloned repository
    shutil.rmtree(temp_repo_path)
"""
