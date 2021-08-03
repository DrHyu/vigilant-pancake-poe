
from bs4 import BeautifulSoup
import json
import requests


def getItemMods():

    URL = 'https://poedb.tw/us/mod.php?l=1'
    data = requests.get(URL)
    soup = BeautifulSoup(data.text, 'html.parser')

    trs = soup.select('#Modsmod_item_list > div > table > tbody > tr')

    mods = {}
    for tr in trs:
        raw_mod = tr.select_one("td:nth-child(3)")

        # Remove the spans containg the numbers
        for span in raw_mod.find_all('span'):
            span.string = "##"

        mods[raw_mod.get_text()] = True

    sorted_mods = sorted(mods)

    print("Found {} mods.".format(len(trs)))
    print("Found {} unique mods.".format(len(sorted_mods)))

    return sorted_mods


def getUniqueNames():
    URL = 'https://poedb.tw/us/Uniques'
    data = requests.get(URL)
    soup = BeautifulSoup(data.text, 'html.parser')

    uniques_names = [unique.get_text() for unique in soup.select(
        '#Uniqueunique_listtitleWeapon > div > table > tbody > tr > td:nth-child(2) > a')]

    uniques_names += [unique.get_text() for unique in soup.select(
        '#Uniqueunique_listtitleArmour > div > table > tbody > tr > td:nth-child(2) > a')]

    uniques_names += [unique.get_text() for unique in soup.select(
        '#Uniqueunique_listtitleOther > div > table > tbody > tr > td:nth-child(2) > a')]

    print("Found {} uniques.".format(len(uniques_names)))

    return uniques_names


def main():

    mods = getItemMods()
    uniques = getUniqueNames()

    jsn = {'names': uniques, 'mods': mods}

    with open('autocompletes.json', 'w', encoding='utf-8') as f:
        f.write(json.dumps(jsn, indent=' '))


if __name__ == '__main__':
    main()
