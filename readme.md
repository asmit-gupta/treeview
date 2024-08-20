## Installation for Windows

1. Download `treeview.exe` from the [Releases](https://github.com/asmit-gupta/treeview/releases) page.
2. Create a folder for TreeView:
   - Open File Explorer
   - Navigate to `C:\`
   - Create a new folder named `TreeView`
3. Move `treeview.exe` to `C:\TreeView\`
4. Add TreeView to PATH:
   - Right-click on 'This PC' or 'My Computer' on the desktop or in File Explorer
   - Click 'Properties'
   - Click 'Advanced system settings'
   - Click 'Environment Variables'
   - Under 'System variables', find and select 'Path', then click 'Edit'
   - Click 'New'
   - Type `C:\TreeView` and press Enter
   - Click 'OK' on all windows to close them
5. Restart any open Command Prompt windows for the changes to take effect.

## Usage

TreeView supports two main commands: `print-dir` and `create-md`. You can use these commands from any location on your PC by opening a Command Prompt and typing:

### Print Directory Command
treeview print-dir C:\path\to\directory

This command prints detailed directory information to the console, including a tree view of the directory structure and statistics about file types and lines of code.

### Create Markdown Command
treeview create-md C:\path\to\directory

This command creates a markdown file with directory details, including the directory tree and statistics. The markdown file will be created in the current directory.

### Flags

TreeView supports the following flag:

- `--respect-gitignore`: When this flag is used, TreeView will respect .gitignore rules and exclude files and directories specified in the .gitignore file.

Example usage with flag:
treeview print-dir C:\path\to\directory --respect-gitignore
or
treeview create-md C:\path\to\directory --respect-gitignore

## Command Details

1. `print-dir`:
   - Displays a colored tree structure of the specified directory
   - Shows summary statistics including total files, directories, size, and lines of code
   - Provides a breakdown of statistics by file extension

2. `create-md`:
   - Creates a markdown file with the directory tree and statistics
   - The markdown file is saved in the current working directory
   - Useful for documentation or sharing directory information

Both commands will display the processing time upon completion.

Remember, you can always use the `--help` flag with any command to get more information:
treeview --help
treeview print-dir --help
treeview create-md --help

These commands will provide you with detailed information about usage and available options.

![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/asmit-gupta/treeview/total)
