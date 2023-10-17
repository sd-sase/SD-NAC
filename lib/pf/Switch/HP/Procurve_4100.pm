package pf::Switch::HP::Procurve_4100;

=head1 NAME

pf::Switch::HP::Procurve_4100 - Object oriented module to access SNMP enabled HP Procurve 4100 switches

=head1 SYNOPSIS

The pf::Switch::HP::Procurve_4100 module implements an object 
oriented interface to access SNMP enabled HP Procurve 4100 switches.

=head1 SNMP

This switch can parse SNMP traps and change a VLAN on a switch port using SNMP.

=cut

use strict;
use warnings;
use Net::SNMP;

use base ('pf::Switch::HP::AOS_Switch_v16.8');

sub description { 'HP ProCurve 4100 Series' }

=head1 AUTHOR

Inverse inc. <info@inverse.ca>

=head1 COPYRIGHT

Copyright (C) 2005-2023 Inverse inc.

=head1 LICENSE

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301,
USA.

=cut

1;

# vim: set shiftwidth=4:
# vim: set expandtab:
# vim: set backspace=indent,eol,start:
