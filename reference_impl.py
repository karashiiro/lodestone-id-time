from datetime import datetime
from math import sin, pow

def lodestone_id_time(id):
    """
    Calculates the creation time of a FFXIV character,
    and returns it as a datetime object.
    """
    if id <= 5000000:
        excel_time = 37.44 / 5000000 * id + 41539.93
    elif id > 28208601:
        excel_time = 305.01 / 4775200 * id + 42030.57
    else:
        excel_time = 4.10315437 * pow(10, 4) \
        + 1.00993557 * pow(10, -4) * id \
        + 31.5417054 * sin(8.57105764 * pow(10, -7) * id)
    unix_time = (excel_time - 25569) * 86400 # Seconds
    return datetime.fromtimestamp(unix_time)
