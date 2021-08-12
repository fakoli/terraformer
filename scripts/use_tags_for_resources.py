#!/usr/bin/env python
from typing import Any, Dict, List

import argparse
import json
import logging
import os
import sys


log = logging.getLogger(__name__)
log.setLevel(logging.INFO)


def sanitize(name):
    safeChars = ("abcdefghijklmnopqrstuvwxyz"
                 "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_ ")
    safe_names = ''.join(
        c for c in name if c in safeChars).replace(
            "_", "-").replace(" ", "_").lower()
    return safe_names


class ChangeResourceNames(object):
    def __init__(self, input_file: str, output_file: str,
                 target_services: List = []) -> None:
        self.input_file: str = input_file
        self.output_file: str = output_file
        self.target_services: List = target_services
        self.services: Dict(str, str) = {}

    def run(self):
        try:
            with open(self.input_file, 'r') as f:
                data = json.load(f)
        except Exception as e:
            log.error(e)
            sys.exit(1)

        resources = data.get("ImportedResource")
        if not resources:
            Exception("The terraformer json file does not have an"
                      "ImportedResource section.")

        for service_name in resources.keys():
            log.info(f"Operating on {service_name}")

            if (service_name in self.target_services) or (
                    not self.target_services):

                data["ImportedResource"][service_name] = self.parse(
                    resources[service_name])

        self.write(data)

    def parse(self, resources: List) -> List[Dict[str, Any]]:
        resource_index = {}
        for i in range(0, len(resources)):
            _type = resources[i]["InstanceInfo"]["Type"]
            attributes = resources[i]["InstanceState"]["attributes"]

            if attributes.get("tags.Name"):
                name = sanitize(attributes["tags.Name"])

                # If the tag name is already in the dictionary, then we need to
                # add the resource to the list of resources that already have
                # that tag name.
                if name not in resource_index:
                    resource_index[name] = -1
                resource_index[name] += 1

                tag_name = name + "_" + str(resource_index[name])
                type_tag_name = "%s.%s" % (
                    _type, tag_name)

                #  Override the name and Id with tag name
                #  This is to avoid conflicts when the same resource is
                #  imported.
                resources[i]["ResourceName"] = tag_name
                resources[i]["InstanceInfo"]["Id"] = type_tag_name
        return resources

    def write(self, data):
        with open(self.output_file, 'w') as f:
            json.dump(data, f, indent=4)


def parse_args(args):
    parser = argparse.ArgumentParser(description='Change the resource names in a terraformer json file.')
    parser.add_argument('--input', '-i', type=str, required=True,
                        help='The terraformer json file.')
    parser.add_argument('--output', '-o', type=str, required=False,
                        help='The output of modified json file.')
    parser.add_argument('--target_services', '-t', type=list, required=False,
                        help='The new resource name.', default=[])
    parsed_args = parser.parse_args(args)

    parse_args.input = os.path.expanduser(parsed_args.input)

    if not parsed_args.output:
        parsed_args.output = os.path.dirname(parsed_args.input) + "/outplan.json"

    return parsed_args


def main():
    args = parse_args(sys.argv[1:])
    ChangeResourceNames(
        args.input,
        args.output,
        args.target_services
    ).run()


if __name__ == "__main__":
    main()
