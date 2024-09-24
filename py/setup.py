from setuptools import setup, find_packages

# Setup script
setup(
    name='pstash',
    version='0.1.4',
    packages=find_packages(),
    python_requires='>=3.6, <4',
    include_package_data=True,
    package_data={
        'pstash': ['_libs/**/*'],  # Include all files under pstash/_libs
    },
    author='Francesco Ink',
    author_email='me@francesco.ink',
    description='P Stash provides encrypted storage and data exchange for Python applications.',
    long_description=open('README.md').read(),
    long_description_content_type='text/markdown',
    url='https://github.com/stregato/stash',
    classifiers=[
        'Programming Language :: Python :: 3',
        'Programming Language :: Python :: 3.6',
        'Programming Language :: Python :: 3.7',
        'Programming Language :: Python :: 3.8',
        'Programming Language :: Python :: 3.9',
        'Programming Language :: Python :: 3.10',
        'Programming Language :: Python :: 3.11',
        'Programming Language :: Python :: 3.12',
        'Operating System :: MacOS :: MacOS X',
        'Operating System :: POSIX :: Linux',
        'Operating System :: Microsoft :: Windows',
    ],
)