set -e
echo "This scripts will create a crontab entry, set to run ghtlogger at midnight \
on mondays and fridays"
read -p "Enter repository in owner/repo format: " repo
read -p "Enter path to gh_token, or leave empty to use env GH_TOKEN instead: " token
read -p "Enter path for csv log, or leave empty to use default [~/owner_repo_traffic.csv]: " csv

# You can change crontab settings as you please below
crontab_entry="0 0 * * 1,5"

[ -z "$token" ] || token="-t $token"
[ -z "$csv" ] || csv="-c $csv"

path="$(dirname -- "${BASH_SOURCE[0]}")" # relative
path="$(cd -- "$path" && pwd)" # absolute
cmd="$path/ghtlogger $repo $token $csv"
echo "Scripts is: $cmd"
read -r -p "Is that correct? [y/N] " response

if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]
then
   echo "Adding crontab entry..."
   (crontab -l; echo "$crontab_entry $cmd") | crontab -
   echo "Done"
else
    echo "Quitting"
    exit
fi
